// Package main implements a client for Greeter service.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var loadBalancerConnection videoconverter.VideoConverterLoadBalancerClient

func main() {
	// Set up a connection to the server.

	if len(os.Args) != 3 {
		println(errors.New("invalid command line arguments, use ./worker {api-ip} {api-port}").Error())
		return
	}
	ip := os.Args[1]
	port := os.Args[2]
	address := ip + ":" + port
	println("trying to connect to: ", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(time.Second*3))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	//apiConnection := api_gateway.NewAPIGateWayClient(conn)
	loadBalancerConnection = videoconverter.NewVideoConverterLoadBalancerClient(conn)

	outputExtension := "mkv"
	storageClient := video_converter.CreateStorageClient()
	storageClient.DownloadSampleVideos()
	for {
		time.Sleep(time.Second * 5)
		//err := connectToCurrentLoadBalancer(apiConnection)
		token, err := upload("video.mp4")
		if err != nil {
			return
			continue
		}
		err = requestConversion(token, outputExtension)
		if err != nil {
			return
			continue
		}
		loopUntilConverted(token)
		if err != nil {
			return
			continue
		}
		err = download(token, outputExtension)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		loadBalancerConnection.MarkTokenAsComplete(ctx, &videoconverter.MarkTokenAsCompleteRequest{Token: token})
		if err == nil {
			break
		}
	}
	println("done uploading, converting and downloading videos")

}

func connectToCurrentLoadBalancer(apiConnection api_gateway.APIGateWayClient) error {
	ctx := context.Background()
	loadbalancer, err := apiConnection.GetLifeGuardCoordinator(ctx, &api_gateway.GetLifeGuardCoordinatorRequest{})
	address := loadbalancer.Ip + ":" + strconv.Itoa(int(loadbalancer.Port))
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(time.Second*3))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	loadBalancerConnection = videoconverter.NewVideoConverterLoadBalancerClient(conn)
	return err
}

func loopUntilConverted(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	status, err := loadBalancerConnection.ConversionStatus(ctx, &videoconverter.ConversionStatusRequest{StatusId: token})
	if err != nil {
		println(" conv check err: ", err.Error())
		return err
	}
	print("in progres..")
	for status.Code != videoconverter.ConversionStatusCode_Done {
		print(".")
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
		status, err = loadBalancerConnection.ConversionStatus(ctx, &videoconverter.ConversionStatusRequest{StatusId: token})
		if err != nil {
			println("conv check err: ", err.Error())
			return err
		}
		time.Sleep(time.Second * 2)
	}
	return nil
}

func requestConversion(token string, outputType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	println("starting conversion for ", token)
	_, err := loadBalancerConnection.StartConversion(ctx, &videoconverter.ConversionRequest{Token: token, InputType: "mp4", OutputType: outputType})
	if err != nil {
		println(err.Error())
		return err
	}
	println("conversion started")
	return nil
}

func upload(fileName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	stream, err := loadBalancerConnection.Upload(ctx)

	if err != nil {
		log.Println("cannot upload image: ", err)
		return "", err
	}

	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token, err := loadBalancerConnection.RequestUploadToken(ctx2, &videoconverter.UploadTokenRequest{})
	if err != nil {
		println(err)
		return "", err
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
		return "", nil
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("cannot read chunk to buffer: ", err)
			return "", nil
		}

		req := &videoconverter.Chunk{
			RequestType: &videoconverter.Chunk_Content{Content: buffer[:n]},
		}

		err = stream.Send(req)
		if err != nil {
			log.Println("cannot send chunk to server: ", err)
			return "", nil
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("cannot receive response: ", err)
		return "", nil
	}

	log.Printf("image uploaded with id: %s, size: %d", res.RetrievalToken)

	return res.RetrievalToken, nil
}

func download(token string, extension string) error {
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
			return errors.New("Download: " + err.Error())
		}

		buf.Write(data.GetContent())
	}

	f, err := os.Create(constants.LocalStorage + "downloaded" + "." + extension)
	if err != nil {
		log.Println("Download, create file: %v", err)
		return errors.New("Download, create file: " + err.Error())
	}

	_, err = f.Write(buf.Bytes())
	if err != nil {
		log.Println("Download, write to file: %v", err)
		return errors.New("Download, write to file: " + err.Error())
	}

	f.Close()
	return nil
}
