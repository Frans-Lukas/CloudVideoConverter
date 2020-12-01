package main

import (
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	video_converter.ImplicitAuth(constants.ConvertedVideosBucketName)
	client := video_converter.CreateStorageClient()
	files, err := ioutil.ReadDir(constants.LocalStorage)
	if err != nil {
		log.Fatal(err)
	}
	filename := "sdf"
	for _, f := range files {
		filename = f.Name()
		break
	}
	client.UploadConvertedPart(filename)
	client.DownloadSpecificParts(strings.Split(filename, "-")[0])
}
