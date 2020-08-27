package controllers

import (
	"encoding/json"
	"log"
	"virt-webui/models"

	"github.com/astaxie/beego"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "kubevirt.io/client-go/api/v1"
)

// Operations about virtual machine
type VMController struct {
	beego.Controller
}

type JsonRequestVMName struct {
	Name string
}

// @Title List VM
// @Description List all virtual machines.
// @Success 200 {object} controllers.JsonResponseListVMSuccess
// @Failure 500 Failed to list VMs.
// @router / [get]
func (v *VMController) GetAll() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		ResponseNotAvaliable(v)
		return
	}

	// Fetch list of VMs
	vmList, err := (*virtClient).VirtualMachine(*namespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vmi list: %v\n", err)
		v.ServeJSON()
		return
	}

	var vms []models.VM
	for _, vm := range vmList.Items {
		var size int
		if vm.Spec.Template.Spec.Domain.CPU.Cores == 1 {
			size = 0
		} else {
			size = 1
		}
		var ready string
		if vm.Status.Ready {
			ready = "Ready"
		} else {
			ready = "Not Ready"
		}
		vms = append(vms, models.VM{vm.Name, vm.Namespace, size, ready})
	}
	v.Data["json"] = JsonResponseListVMSuccess{200, "VMs list success.", vms}
	v.ServeJSON()
}

type JsonResponseListVMSuccess struct {
	StatusCode int
	Message    string
	VMs        []models.VM
}

// @Title Get VM
// @Description Get overview of an exist virtual machine.
// @Param	VMName		path 	string	true		"The vm you want to get"
// @Success 200 {object} controllers.JsonResponseGetVMSuccess
// @Failure 500 Failed to get VM.
// @router /:VMName [get]
func (v *VMController) Get() {
	vmName := v.Ctx.Input.Param(":VMName")
	if vmName == "1" {
		v.Data["json"] = JsonResponseGetVMSuccess{200, vmName + " get success.",
			JsonResponseGetVMSuccessVMInfo{vmName, "image1", 0, "Running", "This is YAML.", "This is Log."}}
	} else {
		v.Data["json"] = JsonResponseBasic{500, "Failed to get " + vmName + "."}
	}
	v.ServeJSON()
}

type JsonResponseGetVMSuccess struct {
	StatusCode int
	Message    string
	VM         JsonResponseGetVMSuccessVMInfo
}

type JsonResponseGetVMSuccessVMInfo struct {
	Name   string
	Image  string
	Size   int
	Status string
	YAML   string
	Log    string
}

// @Title Start VM
// @Description Start an exist virtual machine.
// @Param	body	body	controllers.JsonRequestVMName	true	"The vm you want to start"
// @Success 200 {object} controllers.JsonResponseBasic
// @Failure 500 Failed to start VM.
// @router /start [POST]
func (v *VMController) Start() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		ResponseNotAvaliable(v)
		return
	}

	var jsonReq JsonRequestVMName
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	vmName := jsonReq.Name

	err := (*virtClient).VirtualMachine(*namespace).Start(vmName)
	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " start success."}
	} else {
		v.Data["json"] = JsonResponseBasic{500, "Failed to start " + vmName + "."}
	}
	v.ServeJSON()
}

// @Title Stop VM
// @Description Stop an exist virtual machine.
// @Param	body	body	controllers.JsonRequestVMName	true	"The vm you want to stop"
// @Success 200 {object} controllers.JsonResponseBasic
// @Failure 500 Failed to stop VM.
// @router /stop [POST]
func (v *VMController) Stop() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		ResponseNotAvaliable(v)
		return
	}

	var jsonReq JsonRequestVMName
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	vmName := jsonReq.Name

	err := (*virtClient).VirtualMachine(*namespace).Stop(vmName)
	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " stop success."}
	} else {
		v.Data["json"] = JsonResponseBasic{500, "Failed to stop " + vmName + "."}
	}
	v.ServeJSON()
}

// @Title Create VM
// @Description Create a new virtual machines.
// @Param	body	body	controllers.JsonRequestCreateVM	true	"The VM content"
// @Success 200 {object} controllers.JsonResponseCreateVM
// @Failure 500 Failed to create VM.
// @router / [POST]
func (v *VMController) Create() {
	var jsonReq JsonRequestCreateVM
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	if jsonReq.Name == "1" {
		v.Data["json"] = JsonResponseCreateVM{200, jsonReq.Name + " create success.",
			JsonRequestCreateVM{jsonReq.Name, jsonReq.Image, jsonReq.Size}}
	} else {
		v.Data["json"] = JsonResponseBasic{500, "Failed to create " + jsonReq.Name + "."}
	}
	v.ServeJSON()
}

type JsonRequestCreateVM struct {
	Name  string
	Image string
	Size  int
}

type JsonResponseCreateVM struct {
	StatusCode int
	Message    string
	VM         JsonRequestCreateVM
}

// @Title Rename VM
// @Description Rename an exist virtual machine.
// @Param	VMName	path 	string	true		"The VM you want to rename"
// @Param	body	body	controllers.JsonRequestRename	true	"The new name"
// @Success 200 {object} controllers.JsonResponseRenameSuccess
// @Failure 500 Failed to rename VM.
// @router /:VMName [put]
func (v *VMController) Put() {
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		ResponseNotAvaliable(v)
		return
	}

	vmName := v.Ctx.Input.Param(":VMName")
	var jsonReq JsonRequestRename
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	newName := jsonReq.NewName

	err := (*virtClient).VirtualMachine(*namespace).Rename(vmName, &v1.RenameOptions{NewName: newName})
	if err == nil {
		v.Data["json"] = JsonResponseRenameSuccess{200, "Rename " + vmName + " to " + newName + " success.", newName}
	} else {
		v.Data["json"] = JsonResponseBasic{500, "Failed to rename " + vmName + " to " + newName + "."}
	}
	v.ServeJSON()
}

// @Title Delete VM
// @Description Delete an exist virtual machine.
// @Param	VMName	path	string	true	"The VM you want to delete"
// @Success 200 {object} controllers.JsonResponseBasic
// @Failure 500 Failed to delete VM.
// @router /:VMName [delete]
func (v *VMController) Delete() {
	ok, namespace, dynamicClient := GetDynamicClient()
	if !ok {
		ResponseNotAvaliable(v)
		return
	}

	vmName := v.Ctx.Input.Param(":VMName")
	vmRes := schema.GroupVersionResource{Group: "kubevirt.io",
		Version: "v1alpha3", Resource: "virtualmachines"}
	deletePolicy := k8smetav1.DeletePropagationForeground
	deleteOptions := k8smetav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err := (*dynamicClient).Resource(vmRes).Namespace(*namespace).Delete(vmName, &deleteOptions)

	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " delete success."}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to delete " + vmName + "."}
	}
	v.ServeJSON()
}
