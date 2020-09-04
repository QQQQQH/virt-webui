const VMs_template = {
    template: "#vms_template",
    data() {
        return {
            vmList: [],
            createVMName: '',
            createVMImage: '',
            createVMSize: 0,
            vmToDelete: ''
        }
    },
    methods: {
        getVMs: function () {
            axios.get("/v1/vms/").then((response) => {
                console.log(response)
                this.vmList = response.data.VMs
            }, (err) => {
                console.log(err)
            })
        },
        setMenuOption: function () {
            this.$parent.selectOption(2)
        },
        createVM: function () {
            console.log(this.createVMName)
            console.log(this.createVMImage)
            console.log(this.createVMSize)
            var data = {
                "Name": this.createVMName,
                "Image": this.createVMImage,
                "Size": this.createVMSize,
            }
            axios.post("/v1/vms/", data).then((res) => {
                console.log(res)
                location.reload()
            }, (err) => {
                console.log(err)
            })
        },
        setVMToDelete: function (index) {
            console.log("delete index: ", index)
            this.vmToDelete = this.vmList[index].Name
        },
        deleteVM: function () {
            console.log("delete: " + this.vmToDelete)
            axios.delete("/v1/vms/" + this.vmToDelete).then((res) => {
                console.log(res)
                location.reload()
            }, (err) => {
                console,
                log(err)
            })
        },
        startVM: function (index) {
            var vm = this.vmList[index].Name
            console.log("start: " + vm)
            var data = {
                "Name": vm
            }
            axios.post("/v1/vms/start/", data).then((res) => {
                console.log(res)
            }), (err) => {
                console.log(err)
            }
        },
        stopVM: function (index) {
            var vm = this.vmList[index].Name
            console.log("stop: " + vm)
            var data = {
                "Name": vm
            }
            axios.post("/v1/vms/stop/", data).then((res) => {
                console.log(res)
            }), (err) => {
                console.log(err)
            }
        }
    },
    mounted() {
        this.setMenuOption()
        this.getVMs()
    }
}