package imageupload

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	pb "gopkg.in/cheggaaa/pb.v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	cdiClientset "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned"
	"kubevirt.io/client-go/kubecli"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	uploadcdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/upload/v1alpha1"
)

const (
	// PodPhaseAnnotation is the annotation on a PVC containing the upload pod phase
	PodPhaseAnnotation = "cdi.kubevirt.io/storage.pod.phase"

	// PodReadyAnnotation tells whether the uploadserver pod is ready
	PodReadyAnnotation = "cdi.kubevirt.io/storage.pod.ready"

	uploadRequestAnnotation = "cdi.kubevirt.io/storage.upload.target"

	uploadReadyWaitInterval = 2 * time.Second

	processingWaitInterval = 2 * time.Second
	processingWaitTotal    = 24 * time.Hour

	//UploadProxyURIAsync is a URI of the upload proxy, the endpoint is asynchronous
	UploadProxyURIAsync = "/v1alpha1/upload-async"

	//UploadProxyURI is a URI of the upload proxy, the endpoint is synchronous for backwards compatibility
	UploadProxyURI = "/v1alpha1/upload"

	configName = "config"
)

var (
	insecure       bool
	uploadProxyURL string
	name           string
	size           string
	imagePath      string
	accessMode     string

	uploadPodWaitSecs uint
)

type HTTPClientCreator func(bool) *http.Client

var httpClientCreatorFunc HTTPClientCreator

type processingCompleteFunc func(kubernetes.Interface, string, string, time.Duration, time.Duration) error

// UploadProcessingCompleteFunc the function called while determining if post transfer processing is complete.
var UploadProcessingCompleteFunc processingCompleteFunc = waitUploadProcessingComplete

// SetHTTPClientCreator allows overriding the default http client
// useful for unit tests
func SetHTTPClientCreator(f HTTPClientCreator) {
	httpClientCreatorFunc = f
}

// SetDefaultHTTPClientCreator sets the http client creator back to default
func SetDefaultHTTPClientCreator() {
	httpClientCreatorFunc = getHTTPClient
}

func init() {
	SetDefaultHTTPClientCreator()
}

func UploadImage(insecure0 bool, uploadProxyURL0, name0, size0, imagePath0, accessMode0 string, uploadPodWaitSecs0 uint) error {
	insecure = insecure0
	uploadProxyURL, name, size, imagePath, accessMode = uploadProxyURL0, name0, size0, imagePath0, accessMode0
	uploadPodWaitSecs = uploadPodWaitSecs0
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return err
	}

	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("cannot obtain KubeVirt client: %v", err)
	}

	err = getAndValidateUploadPVC(virtClient, namespace, name)
	if err != nil {
		return err
	}
	dv, err := createUploadDataVolume(virtClient, namespace, name, size, accessMode)
	if err != nil {
		return err
	}

	fmt.Printf("DataVolume %s/%s created\n", dv.Namespace, dv.Name)

	err = waitUploadServerReady(virtClient, namespace, name, uploadReadyWaitInterval, time.Duration(uploadPodWaitSecs)*time.Second)
	if err != nil {
		return err
	}

	if uploadProxyURL == "" {
		uploadProxyURL, err = getUploadProxyURL(virtClient.CdiClient())
		if err != nil {
			return err
		}
		if uploadProxyURL == "" {
			return fmt.Errorf("uploadproxy URL not found")
		}
	}

	u, err := url.Parse(uploadProxyURL)
	if err != nil {
		return err
	}

	if u.Scheme == "" {
		uploadProxyURL = fmt.Sprintf("https://%s", uploadProxyURL)
	}

	fmt.Printf("Uploading data to %s\n", uploadProxyURL)

	token, err := getUploadToken(virtClient.CdiClient(), namespace, name)
	if err != nil {
		return err
	}

	err = uploadData(uploadProxyURL, token, file, insecure)
	if err != nil {
		return err
	}

	fmt.Println("Uploading data completed successfully, waiting for processing to complete, you can hit ctrl-c without interrupting the progress")
	err = UploadProcessingCompleteFunc(virtClient, namespace, name, processingWaitInterval, processingWaitTotal)
	if err != nil {
		fmt.Printf("Timed out waiting for post upload processing to complete, please check upload pod status for progress\n")
	} else {
		fmt.Printf("Uploading %s completed successfully\n", imagePath)
	}

	return err
}

func getAndValidateUploadPVC(client kubernetes.Interface, namespace, name string) error {
	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	return fmt.Errorf("PVC %s already exists.", name)
}

func createUploadDataVolume(client kubecli.KubevirtClient, namespace, name, size, accessMode string) (*cdiv1.DataVolume, error) {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("validation failed for size=%s: %s", size, err)
	}

	dv := &cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: cdiv1.DataVolumeSource{
				Upload: &cdiv1.DataVolumeSourceUpload{},
			},
			PVC: &v1.PersistentVolumeClaimSpec{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: quantity,
					},
				},
			},
		},
	}

	dv.Spec.PVC.AccessModes = []v1.PersistentVolumeAccessMode{v1.PersistentVolumeAccessMode(accessMode)}

	dv, err = client.CdiClient().CdiV1alpha1().DataVolumes(namespace).Create(dv)
	if err != nil {
		return nil, err
	}

	return dv, nil
}

func waitUploadServerReady(client kubernetes.Interface, namespace, name string, interval, timeout time.Duration) error {
	loggedStatus := false

	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			// DataVolume controller may not have created the PVC yet
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		// upload controler sets this to true when uploadserver pod is ready to receive data
		podReady := pvc.Annotations[PodReadyAnnotation]
		done, _ := strconv.ParseBool(podReady)

		if !done && !loggedStatus {
			fmt.Printf("Waiting for PVC %s upload pod to be ready...\n", name)
			loggedStatus = true
		}

		if done && loggedStatus {
			fmt.Printf("Pod now ready\n")
		}

		return done, nil
	})

	return err
}

func getUploadProxyURL(client cdiClientset.Interface) (string, error) {
	cdiConfig, err := client.CdiV1alpha1().CDIConfigs().Get(configName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if cdiConfig.Spec.UploadProxyURLOverride != nil {
		return *cdiConfig.Spec.UploadProxyURLOverride, nil
	}
	if cdiConfig.Status.UploadProxyURL != nil {
		return *cdiConfig.Status.UploadProxyURL, nil
	}
	return "", nil
}

func getUploadToken(client cdiClientset.Interface, namespace, name string) (string, error) {
	request := &uploadcdiv1.UploadTokenRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "token-for-virtctl",
		},
		Spec: uploadcdiv1.UploadTokenRequestSpec{
			PvcName: name,
		},
	}

	response, err := client.UploadV1alpha1().UploadTokenRequests(namespace).Create(request)
	if err != nil {
		return "", err
	}

	return response.Status.Token, nil
}

func uploadData(uploadProxyURL, token string, file *os.File, insecure bool) error {
	url, err := ConstructUploadProxyPathAsync(uploadProxyURL, token, insecure)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	bar := pb.New64(fi.Size()).SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(file)

	client := httpClientCreatorFunc(insecure)
	req, _ := http.NewRequest("POST", url, reader)

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/octet-stream")
	req.ContentLength = fi.Size()

	fmt.Println()
	bar.Start()

	resp, err := client.Do(req)

	bar.Finish()
	fmt.Println()

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected return value %d, %s", resp.StatusCode, string(body))
	}

	return nil
}

//ConstructUploadProxyPathAsync - receives uploadproxy adress and concatenates to it URI
func ConstructUploadProxyPathAsync(uploadProxyURL, token string, insecure bool) (string, error) {
	u, err := url.Parse(uploadProxyURL)

	if err != nil {
		return "", err
	}

	if !strings.Contains(uploadProxyURL, UploadProxyURIAsync) {
		u.Path = path.Join(u.Path, UploadProxyURIAsync)
	}

	// Attempt to discover async URL
	client := httpClientCreatorFunc(insecure)
	req, _ := http.NewRequest("HEAD", u.String(), nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Async not available, use regular upload url.
		return ConstructUploadProxyPath(uploadProxyURL)
	}

	return u.String(), nil
}

//ConstructUploadProxyPath - receives uploadproxy adress and concatenates to it URI
func ConstructUploadProxyPath(uploadProxyURL string) (string, error) {
	u, err := url.Parse(uploadProxyURL)

	if err != nil {
		return "", err
	}

	if !strings.Contains(uploadProxyURL, UploadProxyURI) {
		u.Path = path.Join(u.Path, UploadProxyURI)
	}
	return u.String(), nil
}

func waitUploadProcessingComplete(client kubernetes.Interface, namespace, name string, interval, timeout time.Duration) error {
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// upload controler sets this to true when uploadserver pod is ready to receive data
		podPhase := pvc.Annotations[PodPhaseAnnotation]

		if podPhase == string(v1.PodSucceeded) {
			fmt.Printf("Processing completed successfully\n")
		}

		return podPhase == string(v1.PodSucceeded), nil
	})

	return err
}

func getHTTPClient(insecure bool) *http.Client {
	client := &http.Client{}

	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}
