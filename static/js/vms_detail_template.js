const VMs_detail_template = {
    template: "#vms_detail_template",
    data() {
        return {
            vm: {}
        }
    },
    methods: {
        getVM: function () {
            axios.get("/v1/vms/" + this.$route.params.name).then((response) => {
                console.log(response)
                this.vm = response.data
            }, (err) => {
                console.log(err)
            })
        },
        setMenuOption: function () {
            this.$parent.selectOption(2)
        },
        goBack: function () {
            this.$router.go(-1)
        }
    },
    mounted() {
        this.setMenuOption()
        this.getVM()
    }
}