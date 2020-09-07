const Overview_template = {
    template: "#overview_template",
    data() {
        return {
            imgList: [],
            vmList: [],
        }
    },
    methods: {
        getImages: function () {
            axios.get("/v1/images/").then((response) => {
                console.log(response)
                this.imgList = response.data.Images
            }, (err) => {
                console.log(err)
            })
        },
        getVMs: function () {
            axios.get("/v1/vms/").then((response) => {
                console.log(response)
                this.vmList = response.data.VMs
            }, (err) => {
                console.log(err)
            })
        },
        setMenuOption: function () {
            this.$parent.selectOption(0)
        }
    },
    mounted() {
        this.setMenuOption()
        this.getImages()
        this.getVMs()
    }
}