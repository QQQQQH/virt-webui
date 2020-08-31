package controllers

import (
	"encoding/json"
	"log"
	"virt-webui/models"

	"github.com/astaxie/beego"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		v.ResponseNotAvaliable()
		return
	}

	// Fetch list of VMs
	vmList, err := (*virtClient).VirtualMachine(*namespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vmi list: %v\n", err)
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to list VMs. " + err.Error()}
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
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		v.ResponseNotAvaliable()
		return
	}

	vmName := v.Ctx.Input.Param(":VMName")
	vm, err := (*virtClient).VirtualMachine(*namespace).Get(vmName, &k8smetav1.GetOptions{})

	if err == nil {
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
		v.Data["json"] = JsonResponseGetVMSuccess{200, vmName + " get success.",
			JsonResponseGetVMSuccessVMInfo{vmName, "This is image", size, ready, "This is YAML.", "This is Log."}}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to get " + vmName + ". " + err.Error()}
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
		v.ResponseNotAvaliable()
		return
	}

	var jsonReq JsonRequestVMName
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	vmName := jsonReq.Name

	err := (*virtClient).VirtualMachine(*namespace).Start(vmName)
	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " start success."}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to start " + vmName + ". " + err.Error()}
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
		v.ResponseNotAvaliable()
		return
	}

	var jsonReq JsonRequestVMName
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	vmName := jsonReq.Name

	err := (*virtClient).VirtualMachine(*namespace).Stop(vmName)
	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " stop success."}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to stop " + vmName + ". " + err.Error()}
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
	// ok, namespace, dynamicClient := GetDynamicClient()
	// if !ok {
	// 	v.ResponseNotAvaliable()
	// 	return
	// }

	// vmRes := schema.GroupVersionResource{Group: "kubevirt.io",
	// 	Version: "v1alpha3", Resource: "virtualmachines"}

	// vm := &unstructured.Unstructured{
	// 	Object: map[string]interface{}{
	// 		"apiVersion": "kubevirt.io/v1alpha3",
	// 		"kind":       "VirtualMachine",
	// 		"metadata": map[string]interface{}{
	// 			"name": vmName,
	// 		},
	// 		"spec": map[string]interface{}{
	// 			"running": false,
	// 			"template": map[string]interface{}{
	// 				"metadata": map[string]interface{}{
	// 					"labels": map[string]interface{}{
	// 						"kubevirt.io/domain": vmName,
	// 					},
	// 				},
	// 				"spec": map[string]interface{}{
	// 					"domain": map[string]interface{}{
	// 						"cpu": map[string]interface{}{
	// 							"cores": cpu,
	// 						},
	// 						"devices": map[string]interface{}{
	// 							"disks": []map[string]interface{}{
	// 								{
	// 									"disk": map[string]interface{}{
	// 										"bus": "virtio",
	// 									},
	// 									"name": "dvdisk",
	// 								},
	// 							},
	// 						},
	// 						"resources": map[string]interface{}{
	// 							"requests": map[string]interface{}{
	// 								"memory": memory,
	// 							},
	// 						},
	// 					},
	// 					"volumes": []map[string]interface{}{
	// 						{
	// 							"name": "dvdisk",
	// 							"dataVolume": map[string]interface{}{
	// 								"name": image,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// _, err := (*dynamicClient).Resource(vmRes).Namespace(*namespace).Create(vm, k8smetav1.CreateOptions{})

	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		v.ResponseNotAvaliable()
		return
	}

	var jsonReq JsonRequestCreateVM
	json.Unmarshal(v.Ctx.Input.RequestBody, &jsonReq)
	vmName := jsonReq.Name
	image := jsonReq.Image
	size := jsonReq.Size
	running := false
	var cpu uint32
	var memory string
	if size == 0 {
		cpu = 1
		memory = "1G"
	} else {
		cpu = 2
		memory = "2G"
	}

	vm := v1.VirtualMachine{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "virtualMachine",
			APIVersion: "kubevirt.io/v1alpha3",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name: vmName,
		},
		Spec: v1.VirtualMachineSpec{
			Running: &running,
			Template: &v1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: k8smetav1.ObjectMeta{
					Labels: map[string]string{
						"kubevirt.io/domain": vmName,
					},
				},
				Spec: v1.VirtualMachineInstanceSpec{
					Domain: v1.DomainSpec{
						CPU: &v1.CPU{
							Cores: cpu,
						},
						Devices: v1.Devices{
							Disks: []v1.Disk{
								{
									Name: "dvdisk",
									DiskDevice: v1.DiskDevice{
										Disk: &v1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
							},
						},
						Resources: v1.ResourceRequirements{
							Requests: k8sv1.ResourceList{
								"memory": resource.MustParse(memory),
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "dvdisk",
							VolumeSource: v1.VolumeSource{
								DataVolume: &v1.DataVolumeSource{
									image,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := (*virtClient).VirtualMachine(*namespace).Create(&vm)

	if err == nil {
		v.Data["json"] = JsonResponseCreateVM{200, vmName + " create success.",
			JsonRequestCreateVM{vmName, image, size}}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to create " + vmName + ". " + err.Error()}
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
		v.ResponseNotAvaliable()
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
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to rename " + vmName + " to " + newName + ". " + err.Error()}
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
	ok, namespace, virtClient := GetVirtClient()
	if !ok {
		v.ResponseNotAvaliable()
		return
	}

	vmName := v.Ctx.Input.Param(":VMName")

	err := (*virtClient).VirtualMachine(*namespace).Delete(vmName, &k8smetav1.DeleteOptions{})

	if err == nil {
		v.Data["json"] = JsonResponseBasic{200, vmName + " delete success."}
	} else {
		v.Ctx.Output.SetStatus(500)
		v.Data["json"] = JsonResponseBasic{500, "Failed to delete " + vmName + ". " + err.Error()}
	}
	v.ServeJSON()
}
