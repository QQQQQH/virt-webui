package controllers

import (
	"encoding/json"
	"flag"
	"log"
	"path/filepath"
	imageupload "virt-webui/controllers/imageUpload"
	"virt-webui/models"

	"github.com/astaxie/beego"
	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kubevirt.io/client-go/kubecli"
)

func (i *ImageController) ResponseNotAvaliable() {
	i.Data["json"] = JsonResponseBasic{500, "Not avaliable."}
	i.ServeJSON()
	return
}

func (v *VMController) ResponseNotAvaliable() {
	v.Data["json"] = JsonResponseBasic{500, "Not avaliable."}
	v.ServeJSON()
	return
}

func GetVirtClient() (bool, *string, *kubecli.KubevirtClient) {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// retrive default namespace.
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		log.Fatalf("error in KubeVirt namespace : %v\n", err)
		return false, nil, nil
	}

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
		return false, nil, nil
	}
	return true, &namespace, &virtClient
}

var kubeconfig *string

func GetDynamicClient() (bool, *string, *dynamic.Interface) {
	if kubeconfig == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
	}

	namespace := "default"

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("error in dynamic config : %v\n", err)
		return false, nil, nil
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("error in dynamic client : %v\n", err)
		return false, nil, nil
	}
	return true, &namespace, &client
}

// Operations about image
type ImageController struct {
	beego.Controller
}

type JsonResponseBasic struct {
	StatusCode int
	Message    string
}

// @Title List Image
// @Description List all images.
// @Success 200 {object} controllers.JsonResponseListImageSuccess
// @Failure 500 Failed to list images.
// @router / [get]
func (i *ImageController) GetAll() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		i.ResponseNotAvaliable()
		return
	}
	imgList, err := (*virtClient).CdiClient().CdiV1alpha1().DataVolumes(*namespace).List(k8smetav1.ListOptions{})

	if err != nil {
		log.Fatalf("cannot obtain KubeVirt image list: %v\n", err)
		i.Ctx.Output.SetStatus(500)
		i.Data["json"] = JsonResponseBasic{500, "Failed to list images. " + err.Error()}
		i.ServeJSON()
		return
	}

	var imgs []models.Image
	for _, img := range imgList.Items {
		imgs = append(imgs, models.Image{img.Name, img.Namespace})
	}

	i.Data["json"] = JsonResponseListImageSuccess{200, "Images list success.", imgs}
	i.ServeJSON()
}

type JsonResponseListImageSuccess struct {
	StatusCode int
	Message    string
	Images     []models.Image
}

// @Title Upload Image
// @Description Upload a new image.
// @Param	body	body	controllers.JsonRequestUploadImage	true	"The image content"
// @Success 200 {object} controllers.JsonResponseUploadImageSuccess
// @Failure 500 Failed to upload image.
// @router / [post]
func (i *ImageController) Post() {
	var jsonReq JsonRequestUploadImage
	json.Unmarshal(i.Ctx.Input.RequestBody, &jsonReq)

	insecure := true
	uploadProxyUrl := jsonReq.UploadProxyUrl
	name := jsonReq.Name
	size := jsonReq.Size
	imagePath := jsonReq.FilePath
	accessMode := "ReadWriteOnce"
	uploadPodWaitSecs := uint(240)

	err := imageupload.UploadImage(insecure, uploadProxyUrl, name, size, imagePath, accessMode, uploadPodWaitSecs)

	if err == nil {
		i.Data["json"] = JsonResponseUploadImageSuccess{200, name + " upload success.",
			JsonRequestUploadImage{name, imagePath, uploadProxyUrl, size}}
	} else {
		i.Ctx.Output.SetStatus(500)
		i.Data["json"] = JsonResponseBasic{500, "Failed to upload " + name + ". " + err.Error()}
	}
	i.ServeJSON()
}

type JsonRequestUploadImage struct {
	Name           string
	FilePath       string
	UploadProxyUrl string
	Size           string
}

type JsonResponseUploadImageSuccess struct {
	StatusCode int
	Message    string
	Image      JsonRequestUploadImage
}

// // @Title Rename Image
// // @Description Rename an exist image.
// // @Param	ImageName	path 	string	true		"The image you want to rename"
// // @Param	body	body	controllers.JsonRequestRename	true	"The new name"
// // @Success 200 {object} controllers.JsonResponseRenameSuccess
// // @Failure 500 Failed to rename image.
// // @router /:ImageName [put]
// func (i *ImageController) Put() {
// 	ok, namespace, virtClient := GetVirtClient()
// 	if !ok {
// 		i.ResponseNotAvaliable()
// 		return
// 	}

// 	imgName := i.Ctx.Input.Param(":ImageName")
// 	var jsonReq JsonRequestRename
// 	json.Unmarshal(i.Ctx.Input.RequestBody, &jsonReq)
// 	newName := jsonReq.NewName

// 	img, err := (*virtClient).CdiClient().CdiV1alpha1().DataVolumes(*namespace).Get(imgName, k8smetav1.GetOptions{})

// 	if err == nil {
// 		newImg := img.DeepCopy()
// 		newImg.SetName(newName)
// 		err = (*virtClient).CdiClient().CdiV1alpha1().DataVolumes(*namespace).Delete(imgName, &k8smetav1.DeleteOptions{})
// 		if err == nil {
// 			_, err = (*virtClient).CdiClient().CdiV1alpha1().DataVolumes(*namespace).Create(newImg)
// 			if err == nil {
// 				i.Data["json"] = JsonResponseRenameSuccess{200, "Rename " + imgName + " to " + newName + " success.", newName}
// 			} else {
// 				fmt.Printf("%s\n", err)
// 			}
// 		}
// 	}
// 	if err != nil {
//		i.Ctx.Output.SetStatus(500)
// 		i.Data["json"] = JsonResponseBasic{500, "Failed to rename " + imgName + " to " + newName + ". " + err.Error()}
// 	}
// 	i.ServeJSON()
// }

type JsonRequestRename struct {
	NewName string
}

type JsonResponseRenameSuccess struct {
	StatusCode int
	Message    string
	NewName    string
}

// @Title Delete Image
// @Description Delete an exist image.
// @Param	ImageName	path	string	true	"The image you want to delete"
// @Success 200 {object} controllers.JsonResponseBasic
// @Failure 500 Failed to delete image.
// @router /:ImageName [delete]
func (i *ImageController) Delete() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		i.ResponseNotAvaliable()
		return
	}

	imgName := i.Ctx.Input.Param(":ImageName")
	err := (*virtClient).CdiClient().CdiV1alpha1().DataVolumes(*namespace).Delete(imgName, &k8smetav1.DeleteOptions{})

	if err == nil {
		i.Data["json"] = JsonResponseBasic{200, imgName + " delete success."}
	} else {
		i.Ctx.Output.SetStatus(500)
		i.Data["json"] = JsonResponseBasic{500, "Failed to delete " + imgName + ". " + err.Error()}
	}
	i.ServeJSON()
}
