package converter

import (
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/server/items"
	"os"
	"os/exec"
	"time"
)

type VideoConverterServiceServer struct {
	videoconverter.UnimplementedVideoConverterServiceServer
	ActiveTokens *map[string]items.Token
}

func CreateNewVideoConverterServiceServer() VideoConverterServiceServer {
	activeTokens := make(map[string]items.Token, 0)
	val := VideoConverterServiceServer{
		ActiveTokens: &activeTokens,
	}
	return val
}

func (serv *VideoConverterServiceServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	creationTime := time.Now()
	isStarted := false
	isDone := false
	isFailed := false
	(*serv.ActiveTokens)[in.Token] = items.Token{
		CreationTime:      &creationTime,
		ConversionStarted: &isStarted,
		ConversionDone:    &isDone,
		ConversionFailed:  &isFailed,
		ConvertTo:         &in.OutputType,
	}

	serv.actuallyStartConversion(in.Token, in.OutputType)

	return &videoconverter.ConversionResponse{}, nil
}

func (serv *VideoConverterServiceServer) performConversion(app string, arg0 string, arg1 string, arg2 string, token string) {
	cmd := exec.Command(app, arg0, arg1, arg2)
	*(*serv.ActiveTokens)[token].ConversionStarted = true
	err := cmd.Run()
	if err != nil {
		*(*serv.ActiveTokens)[token].ConversionFailed = true
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
	os.Rename(arg2, constants.LocalStorage+token)
}

func (serv *VideoConverterServiceServer) ConversionStatus(ctx context.Context, in *videoconverter.AvailableForWorkRequest) (*videoconverter.AvailableForWorkResponse, error) {
	for _, v := range *serv.ActiveTokens {
		if !*v.ConversionDone {
			//TODO decide if inProgress is a good response
			return &videoconverter.AvailableForWorkResponse{AvailableForWork: false}, nil
		}
	}

	//TODO decide if notStarted is a good default response
	return &videoconverter.AvailableForWorkResponse{AvailableForWork: true}, nil
}

func (serv *VideoConverterServiceServer) downloadFileToConvert(token string) {
	//TODO storageClient should not be created each time a download happens
	storageClient := video_converter.CreateStorageClient()
	storageClient.DownloadUnconvertedPart(token)
}

func (serv *VideoConverterServiceServer) HandleConversionsLoop() {
	for {
		for tokenString, token := range *serv.ActiveTokens {
			if *token.ConversionDone {
				serv.uploadConvertedFile(tokenString)
				serv.deleteFiles(tokenString, token)

				//TODO remove token from activeTokens
				continue
			}

			if *token.ConversionFailed {
				serv.actuallyStartConversion(tokenString, *token.ConvertTo)
			}

		}
		time.Sleep(time.Second * 5)
	}
}

func (serv *VideoConverterServiceServer) actuallyStartConversion(token string, outputType string) {
	filePath := constants.LocalStorage + token + ".mp4"

	app := "ffmpeg"
	arg0 := "-i"
	arg1 := filePath
	arg2 := constants.LocalStorage + token + "." + outputType

	go func() {
		serv.downloadFileToConvert(token)
		serv.performConversion(app, arg0, arg1, arg2, token)
	}()
}

func (serv *VideoConverterServiceServer) deleteFiles(tokenString string, token items.Token) {
	filePath := constants.LocalStorage + tokenString
	_, err := os.Stat(filePath)
	if err == nil {
		println("deleting " + filePath)
		err := os.Remove(filePath)
		if err != nil {
			println(err.Error())
		}
	}
}

func (serv *VideoConverterServiceServer) uploadConvertedFile(token string) {
	//TODO storageClient should not be created each time a download happens
	storageClient := video_converter.CreateStorageClient()
	storageClient.UploadConvertedPart(token)
}
