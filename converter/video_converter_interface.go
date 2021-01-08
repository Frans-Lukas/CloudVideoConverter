package converter

import (
	"bytes"
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

const maxNumberOfSimulConversions = 2

type VideoConverterServiceServer struct {
	videoconverter.UnimplementedVideoConverterServiceServer
	ActiveConversions *map[string]items.Token
	thisAddress       string
	databaseClient    *video_converter.ConversionObjectsClient
	storageClient     *video_converter.StorageClient
	name              string
	activeConversions *int
}

func CreateNewVideoConverterServiceServer(address string, name string) VideoConverterServiceServer {
	activeTokens := make(map[string]items.Token, 0)
	dataBaseClient := video_converter.NewConversionObjectsClient()
	storageClient := video_converter.CreateStorageClient()
	numConvs := 0
	val := VideoConverterServiceServer{
		ActiveConversions: &activeTokens,
		databaseClient:    &dataBaseClient,
		storageClient:     &storageClient,
		thisAddress:       address,
		name:              name,
		activeConversions: &numConvs,
	}
	return val
}

func (serv *VideoConverterServiceServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	println("starting conversion for ", in.Token, " to type: ", in.OutputType)
	//serv.databaseClient.
	creationTime := time.Now()
	isStarted := false
	isDone := false
	isFailed := false
	fileName := in.Token
	(*serv.ActiveConversions)[fileName] = items.Token{
		CreationTime:      &creationTime,
		ConversionStarted: &isStarted,
		ConversionDone:    &isDone,
		ConversionFailed:  &isFailed,
		OutputType:        &in.OutputType,
	}

	serv.databaseClient.SetConversionInProgressForPart(fileName, serv.thisAddress)

	serv.actuallyStartConversion(fileName, in.OutputType)

	return &videoconverter.ConversionResponse{}, nil
}

func (serv *VideoConverterServiceServer) performConversion(app string, arg0 string, arg1 string, arg2 string, token string) {
	println("has downloaded file and is ACTUALLY starting conversion")
	println(app, " -y ", arg0, " ", arg1, " ", arg2)
	cmd := exec.Command(app, "-y", arg0, arg1, arg2)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	*(*serv.ActiveConversions)[token].ConversionStarted = true
	if err != nil {
		*(*serv.ActiveConversions)[token].ConversionFailed = true
		log.Println("failed to convert " + out.String())
		log.Println(err.Error(), ": ", stderr.String())
		return
	} else {
		*(*serv.ActiveConversions)[token].ConversionDone = true
		log.Println(out.String())
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
	if *serv.activeConversions < maxNumberOfSimulConversions {
		return &videoconverter.AvailableForWorkResponse{AvailableForWork: true}, nil
	}
	return &videoconverter.AvailableForWorkResponse{AvailableForWork: false}, nil
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
		log.Println("could not shutdown: " + err.Error())
	}
}

func (serv *VideoConverterServiceServer) downloadFileToConvert(token string) {
	serv.storageClient.DownloadUnconvertedPart(token)
}

func (serv *VideoConverterServiceServer) HandleConversionsLoop() {
	for {
		println("handling any converted files.")
		for fileName, token := range *serv.ActiveConversions {
			if *token.ConversionDone {
				correctFileName := helpers.ChangeFileExtension(fileName, *token.OutputType)
				println("uploading ", correctFileName)
				serv.uploadConvertedFile(correctFileName)
				serv.databaseClient.MarkConversionAsDone(fileName)
				serv.storageClient.DeleteUnconvertedPart(fileName)
				DeleteFiles(fileName)
				DeleteFiles(correctFileName)
				delete(*serv.ActiveConversions, fileName)
				continue
			}

			if *token.ConversionFailed {
				serv.actuallyStartConversion(fileName, *token.OutputType)
			}

		}
		video_converter.PrintCPUUsage(serv.name)
		time.Sleep(time.Second * 5)
	}
}

func (serv *VideoConverterServiceServer) actuallyStartConversion(fileName string, outputType string) {
	filePath := constants.LocalStorage + fileName

	app := "ffmpeg"
	arg0 := "-i"
	arg1 := filePath
	arg2 := helpers.ChangeFileExtension(constants.LocalStorage+fileName, outputType)
	*(*serv.ActiveConversions)[fileName].ConversionStarted = true
	go func() {
		*serv.activeConversions++
		defer serv.lowerActiveConversions()
		serv.downloadFileToConvert(fileName)
		serv.performConversion(app, arg0, arg1, arg2, fileName)
	}()
}

func (serv *VideoConverterServiceServer) lowerActiveConversions() {
	*serv.activeConversions--
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
