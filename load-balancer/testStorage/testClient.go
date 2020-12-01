package main

import (
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	println("authenticating...")
	video_converter.ImplicitAuth(constants.ProjectID)
	println("authenticated!")
	println("creating storage client...")

	client := video_converter.CreateStorageClient()
	println("created storage client!")
	println("listing files")
	files, err := ioutil.ReadDir(constants.LocalStorage)
	if err != nil {
		log.Fatal(err)
	}
	filename := "sdf"
	for _, f := range files {
		println(filename)
		filename = f.Name()
		if len(strings.Split(filename, "-")) > 1 {
			println("uploading part: " + filename)
			client.UploadConvertedPart(filename)
			println("uploaded part! ")

		}
	}
	println("downloading part... ")
	client.DownloadSpecificParts("cyD86dl4yHJ91govrQWz")
	println("downloaded part!")
}
