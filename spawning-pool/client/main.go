// Package main implements a client for Greeter service.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

var loadBalancerConnection videoconverter.VideoConverterLoadBalancerClient

func main() {
	// Set up a connection to the server.

	if len(os.Args) != 3 {
		println(errors.New("invalid command line arguments, use ./worker {ip} {port}").Error())
		return
	}
	ip := os.Args[1]
	port := os.Args[2]
	address := ip + ":" + port
	println("trying to connect to: ", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	loadBalancerConnection = videoconverter.NewVideoConverterLoadBalancerClient(conn)

	outputExtension := "mkv"
	storageClient := video_converter.CreateStorageClient()
	storageClient.DownloadSampleVideos()

	token := upload("video.mp4")
	requestConversion(token, outputExtension)
	loopUntilConverted(token)
	download(token, outputExtension)
	println("done uploading, converting and downloading videos")

}

func loopUntilConverted(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	status, err := loadBalancerConnection.ConversionStatus(ctx, &videoconverter.ConversionStatusRequest{StatusId: token})
	if err != nil {
		println(" conv check err: ", err.Error())
	}
	print("in progres..")
	for status.Code != videoconverter.ConversionStatusCode_Done {
		print(".")
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
		status, err = loadBalancerConnection.ConversionStatus(ctx, &videoconverter.ConversionStatusRequest{StatusId: token})
		if err != nil {
			println("conv check err: ", err.Error())
		}
		time.Sleep(time.Second * 2)
	}
}

func requestConversion(token string, outputType string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	println("starting conversion for ", token)
	_, err := loadBalancerConnection.StartConversion(ctx, &videoconverter.ConversionRequest{Token: token, InputType: "mp4", OutputType: outputType})
	if err != nil {
		println(err.Error())
	}
	println("conversion started")
}

func upload(fileName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	stream, err := loadBalancerConnection.Upload(ctx)

	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token, err := loadBalancerConnection.RequestUploadToken(ctx2, &videoconverter.UploadTokenRequest{})
	if err != nil {
		println(err)
		return ""
	}

	req := videoconverter.Chunk{
		RequestType: &videoconverter.Chunk_Token{Token: token.Token},
	}

	stream.Send(&req)

	file, err := os.Open(constants.LocalStorage + fileName)
	defer file.Close()

	v1, _ := os.Getwd()
	println(v1)

	if err != nil {
		println("cannot open file", err.Error())
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot read chunk to buffer: ", err)
		}

		req := &videoconverter.Chunk{
			RequestType: &videoconverter.Chunk_Content{Content: buffer[:n]},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.RetrievalToken)

	return res.RetrievalToken
}

func download(token string, extension string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := videoconverter.DownloadRequest{Id: token}
	stream, err := loadBalancerConnection.Download(ctx, &request)

	buf := bytes.Buffer{}

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Download: %v", err)
		}

		buf.Write(data.GetContent())
	}

	f, err := os.Create(constants.LocalStorage + "downloaded" + "." + extension)
	if err != nil {
		log.Fatalf("Download, create file: %v", err)
	}

	_, err = f.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Download, write to file: %v", err)
	}

	f.Close()
}
