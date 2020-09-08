const Images_template = {
    template: "#images_template",
    data() {
        return {
            imgList: [],
            uploadName: '',
            uploadImagePath: '',
            uploadSize: '1G',
            uploadProxyUrl: 'https://192.168.39.22:31001',
            imgToDelete: '',
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
        setMenuOption: function () {
            this.$parent.selectOption(1)
        },
        uploadImage: function () {
            console.log("upload: ")
            console.log(this.uploadName)
            console.log(this.uploadImagePath)
            console.log(this.uploadSize)
            console.log(this.uploadProxyUrl)
            var data = {
                "Name": this.uploadName,
                "FilePath": this.uploadImagePath,
                "Size": this.uploadSize,
                "UploadProxyUrl": this.uploadProxyUrl
            }
            axios.post("/v1/images/", data).then((res) => {
                console.log(res)
            }, (err) => {
                console.log(err)
            })
        },
        setImgToDelete: function (index) {
            console.log("delete index: ", index)
            this.imgToDelete = this.imgList[index].Name
        },
        deleteImage: function () {
            console.log("delete: " + this.imgToDelete)
            axios.delete("/v1/images/" + this.imgToDelete).then((res) => {
                console.log(res)
                location.reload()
            }, (err) => {
                console,
                log(err)
            })
        }
    },
    mounted() {
        this.setMenuOption()
        this.getImages()
    }
}