const Overview_template = {
    template: "#overview_template",
    methods: {
        setMenuOption: function () {
            this.$parent.selectOption(0)
        }
    },
    mounted() {
        this.setMenuOption()
    }
}