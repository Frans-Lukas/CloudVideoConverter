package converter

import (
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/helpers"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/server/items"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type VideoConverterServiceServer struct {
	videoconverter.UnimplementedVideoConverterServiceServer
	ActiveTokens   *map[string]items.Token
	databaseClient *video_converter.ConversionObjectsClient
	storageClient  *video_converter.StorageClient
}

func CreateNewVideoConverterServiceServer() VideoConverterServiceServer {
	activeTokens := make(map[string]items.Token, 0)
	dataBaseClient := video_converter.NewConversionObjectsClient()
	storageClient := video_converter.CreateStorageClient()
	val := VideoConverterServiceServer{
		ActiveTokens:   &activeTokens,
		databaseClient: &dataBaseClient,
		storageClient:  &storageClient,
	}
	return val
}

func (serv *VideoConverterServiceServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	println("starting conversion for ", in.Token, " to type: ", in.OutputType)
	creationTime := time.Now()
	isStarted := false
	isDone := false
	isFailed := false
	(*serv.ActiveTokens)[in.Token] = items.Token{
		CreationTime:      &creationTime,
		ConversionStarted: &isStarted,
		ConversionDone:    &isDone,
		ConversionFailed:  &isFailed,
		OutputType:        &in.OutputType,
	}

	serv.actuallyStartConversion(in.Token, in.OutputType)

	return &videoconverter.ConversionResponse{}, nil
}

func (serv *VideoConverterServiceServer) performConversion(app string, arg0 string, arg1 string, arg2 string, token string) {
	println("has downloaded file and is ACTUALLY starting conversion")
	println(app, " ", arg0, " ", arg1, " ", arg2)
	cmd := exec.Command(app, arg0, arg1, arg2)
	*(*serv.ActiveTokens)[token].ConversionStarted = true
	err := cmd.Run()
	if err != nil {
		*(*serv.ActiveTokens)[token].ConversionFailed = true
		println("failed to convert ", err.Error())
		return
	} else {
		*(*serv.ActiveTokens)[token].ConversionDone = true
	}
	file, err := os.Open(arg2)
	defer file.Close()
	if err != nil {
		println(err.Error())
		return
	}
	println("done converting, target is: ", arg2)
}

func (serv *VideoConverterServiceServer) AvailableForWork(ctx context.Context, in *videoconverter.AvailableForWorkRequest) (*videoconverter.AvailableForWorkResponse, error) {
	for _, v := range *serv.ActiveTokens {
		if !*v.ConversionDone {
			//TODO decide if inProgress is a good response
			return &videoconverter.AvailableForWorkResponse{AvailableForWork: false}, nil
		}
	}

	//TODO decide if notStarted is a good default response
	return &videoconverter.AvailableForWorkResponse{AvailableForWork: true}, nil
}

func (serv *VideoConverterServiceServer) ShutDown(ctx context.Context, in *videoconverter.ShutDownRequest) (*videoconverter.ShutDownResponse, error) {
	go func() {
		time.Sleep(time.Second)
		shutDown()
	}()
	return &videoconverter.ShutDownResponse{}, nil
}

func (serv *VideoConverterServiceServer) IsAlive(ctx context.Context, in *videoconverter.IsAliveRequest) (*videoconverter.IsAliveResponse, error) {
	return &videoconverter.IsAliveResponse{}, nil
}

func shutDown() {
	cmd := exec.Command("/home/group9/CloudVideoConverter/scripts/VMDeleteSelf.sh")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("could not shutdown: " + err.Error())
	}
}

func (serv *VideoConverterServiceServer) downloadFileToConvert(token string) {
	serv.storageClient.DownloadUnconvertedPart(token)
}

func (serv *VideoConverterServiceServer) HandleConversionsLoop() {
	for {
		println("handling any converted files.")
		for fileName, token := range *serv.ActiveTokens {
			if *token.ConversionDone {
				correctFileName := helpers.ChangeFileExtension(fileName, *token.OutputType)
				println("uploading ", correctFileName)
				serv.uploadConvertedFile(correctFileName)
				serv.databaseClient.MarkConversionAsDone(fileName)
				serv.storageClient.DeleteUnconvertedPart(fileName)
				DeleteFiles(fileName)
				DeleteFiles(correctFileName)
				delete(*serv.ActiveTokens, fileName)
				continue
			}

			if *token.ConversionFailed {
				serv.actuallyStartConversion(fileName, *token.OutputType)
			}

		}
		time.Sleep(time.Second * 5)
	}
}

func (serv *VideoConverterServiceServer) actuallyStartConversion(token string, outputType string) {
	filePath := constants.LocalStorage + token

	app := "ffmpeg"
	arg0 := "-i"
	arg1 := filePath
	arg2 := helpers.ChangeFileExtension(constants.LocalStorage+token, outputType)

	go func() {
		serv.downloadFileToConvert(token)
		serv.performConversion(app, arg0, arg1, arg2, token)
	}()
}

func DeleteFiles(prefix string) {
	files, err := filepath.Glob(constants.LocalStorage + prefix)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			println("could not delete file: ", err)
		}
	}
}

func (serv *VideoConverterServiceServer) uploadConvertedFile(token string) {
	//TODO storageClient should not be created each time a download happens
	storageClient := video_converter.CreateStorageClient()
	storageClient.UploadConvertedPart(token)
}
