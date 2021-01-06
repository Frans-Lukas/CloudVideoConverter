package main

import pb "github.com/Frans-Lukas/cloudvideoconverter/load-balancer"

func main() {
	c := pb.NewConversionObjectsClient()
	c.DeleteAllEntities()
	video_store := pb.CreateStorageClient()
	video_store.DeleteAll()
}
