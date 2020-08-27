package controllers

import (
	"encoding/json"
	"log"
	"virt-webui/models"

	"github.com/astaxie/beego"
	"github.com/spf13/pflag"
	"kubevirt.io/client-go/kubecli"
)

func ResponseNoKubeVirt(v *VMController) {
	v.Data["json"] = JsonResponseBasic{500, "KubeVirt not avaliable."}
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
		log.Fatalf("error in namespace : %v\n", err)
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
	var images []models.Image
	images = append(images, models.Image{"image1"})
	images = append(images, models.Image{"image2"})
	i.Data["json"] = JsonResponseListImageSuccess{200, "Images get success.", images}
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
	if jsonReq.Name == "1" {
		i.Data["json"] = JsonResponseUploadImageSuccess{200, jsonReq.Name + " upload success.", JsonRequestUploadImage{jsonReq.Name, jsonReq.FilePath}}

	} else {
		i.Data["json"] = JsonResponseBasic{500, "Failed to upload " + jsonReq.Name + "."}
	}
	i.ServeJSON()
}

type JsonRequestUploadImage struct {
	Name     string
	FilePath string
}

type JsonResponseUploadImageSuccess struct {
	StatusCode int
	Message    string
	Image      JsonRequestUploadImage
}

// @Title Rename Image
// @Description Rename an exist image.
// @Param	ImageName	path 	string	true		"The image you want to rename"
// @Param	body	body	controllers.JsonRequestRename	true	"The new name"
// @Success 200 {object} controllers.JsonResponseRenameSuccess
// @Failure 500 Failed to rename image.
// @router /:ImageName [put]
func (i *ImageController) Put() {
	imageName := i.Ctx.Input.Param(":ImageName")
	var jsonReq JsonRequestRename
	json.Unmarshal(i.Ctx.Input.RequestBody, &jsonReq)
	newName := jsonReq.NewName
	if imageName == "1" {
		i.Data["json"] = JsonResponseRenameSuccess{200, "Rename " + imageName + " to " + newName + " success.", newName}
	} else {
		i.Data["json"] = JsonResponseBasic{500, "Failed to rename " + imageName + " to " + newName + "."}
	}
	i.ServeJSON()
}

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
	imageName := i.Ctx.Input.Param(":ImageName")
	if imageName == "1" {
		i.Data["json"] = JsonResponseBasic{200, imageName + " delete success."}
	} else {
		i.Data["json"] = JsonResponseBasic{500, "Failed to delete " + imageName + "."}
	}
	i.ServeJSON()
}
