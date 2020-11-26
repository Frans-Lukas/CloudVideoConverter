package video_converter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/server/items"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const tokenLength = 20
const tokenTimeOutSeconds = 60 * 20
const megaByte = 1000000
const sizeLimit = megaByte * 1

type VideoConverterServer struct {
	videoconverter.UnimplementedVideoConverterServer
	ActiveTokens *map[string]items.Token
}

func CreateNewServer() VideoConverterServer {
	activeTokens := make(map[string]items.Token, 0)
	val := VideoConverterServer{
		ActiveTokens: &activeTokens,
	}
	return val
}

func (serv *VideoConverterServer) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	creationTime := time.Now()
	isStarted := false
	isDone := false
	isFailed := false
	(*serv.ActiveTokens)[tokenString] = items.Token{
		CreationTime:      &creationTime,
		ConversionStarted: &isStarted,
		ConversionDone:    &isDone,
		ConversionFailed:  &isFailed,
	}
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func saveImage(fileName string, imageBytes *bytes.Buffer) error {
	imagePath := constants.FileDirectory + fileName + ".mp4"
	file, err := os.Create(imagePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}
	_, err = imageBytes.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write image to file: %w", err)
	}
	return nil
}

func (serv *VideoConverterServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	if serv.tokenIsInvalid(in.Token) {
		return nil, errors.New("token is invalid or has timed out: " + in.Token)
	}

	if serv.conversionIsInProgressForToken(in.Token) {
		return nil, errors.New("conversion is already in progress for token: " + in.Token)
	}

	filePath := constants.FileDirectory + in.Token + ".mp4"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("invalid token")
	}
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	app := "ffmpeg"
	arg0 := "-i"
	arg1 := filePath
	arg2 := constants.FileDirectory + in.Token + "." + in.OutputType
	serv.resetConversionStatus(in.Token)

	go func() {
		serv.performConversion(app, arg0, arg1, arg2, in)
	}()

	return &videoconverter.ConversionResponse{}, nil
}

func (serv *VideoConverterServer) conversionIsInProgressForToken(token string) bool {
	// conversion is started but not done or failed
	return *(*serv.ActiveTokens)[token].ConversionStarted && !*(*serv.ActiveTokens)[token].ConversionDone && !*(*serv.ActiveTokens)[token].ConversionFailed
}

func (serv *VideoConverterServer) performConversion(app string, arg0 string, arg1 string, arg2 string, in *videoconverter.ConversionRequest) {
	cmd := exec.Command(app, arg0, arg1, arg2)
	*(*serv.ActiveTokens)[in.Token].ConversionStarted = true
	err := cmd.Run()
	if err != nil {
		*(*serv.ActiveTokens)[in.Token].ConversionFailed = true
		return
	} else {
		*(*serv.ActiveTokens)[in.Token].ConversionDone = true
	}
	file, err := os.Open(arg2)
	defer file.Close()
	if err != nil {
		println(err.Error())
		return
	}
	os.Rename(arg2, constants.FileDirectory+in.Token)
}

func (serv *VideoConverterServer) Upload(stream videoconverter.VideoConverter_UploadServer) error {

	imageData := bytes.Buffer{}
	tokenString := ""

	for {
		streamData, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			}

			err = errors.New("failed unexpectadely while reading chunks from stream")
			return err
		}
		switch streamData.RequestType.(type) {
		case *videoconverter.Chunk_Content:
			if tokenString == "" {
				return errors.New("token must be first message")
			}
			chunk := streamData.GetContent()
			imageData.Write(chunk)
		case *videoconverter.Chunk_Token:
			token := streamData.GetToken()
			if serv.tokenIsInvalid(token) {
				return errors.New("token is invalid or has timed out: " + token)
			}
			tokenString = token
		}
	}

	// once the transmission finished, send the
	// confirmation if nothign went wrong
	err := stream.SendAndClose(&videoconverter.UploadStatus{
		RetrievalToken: tokenString,
	})

	err = saveImage(tokenString, &imageData)
	if err != nil {
		return err
	}

	splitVideo(tokenString)

	// ...

	return nil
}
func splitVideo(token string) {
	filePath := constants.FileDirectory + token + ".mp4"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf(errors.New("video to split does not exist").Error())
	}
	timeInSeconds, timeInSecondsString := getVideoTimeInSeconds(filePath)
	size := getVideoSize(filePath)
	numberOfSplits := int(math.Round(float64(size)/float64(sizeLimit) + 0.49))
	slizeSize := int(timeInSeconds) / numberOfSplits
	println("number of seconds: " + strconv.Itoa(int(timeInSeconds)))

	startTime := 0
	endTime := slizeSize
	for i := 1; i <= numberOfSplits; i++ {
		// must be string because of potential double inprecision
		println("start: ", startTime)
		println("end: ", endTime)
		performSplit(startTime, strconv.Itoa(endTime), filePath, i, numberOfSplits, token)

		startTime = endTime
		endTime += slizeSize
		if endTime > int(timeInSeconds)-slizeSize/2 {
			performSplit(startTime, timeInSecondsString, filePath, i+1, numberOfSplits, token)
			break
		}
	}

}
func getVideoSize(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("invalid file " + filePath)
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		log.Fatalf("could not get file stats " + filePath)
	}
	return int(fi.Size())
}
func getVideoTimeInSeconds(filePath string) (float64, string) {
	println(filePath)
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf(errors.New("failed to get video time in seconds").Error())
	}
	timeString := out.String()
	timeString = strings.Replace(timeString, "\n", "", -1)
	timeInSec, err := strconv.ParseFloat(timeString, 64)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return timeInSec, timeString
}

func performSplit(startTime int, endTime string, filePath string, index int, total int, token string) error {
	targetFileName := constants.FileDirectory + token + "-" + strconv.Itoa(index) + "_" + strconv.Itoa(total) + ".mp4"
	str := "ffmpeg" + " -ss " + strconv.Itoa(startTime) + " -t " + endTime + " -i " + filePath + " -acodec " + " copy " + " -vcodec " + " copy " + targetFileName
	println(str)
	out, err := exec.Command("ffmpeg", "-ss", strconv.Itoa(startTime), "-t", endTime, "-i", filePath, "-acodec", "copy", "-vcodec", "copy", targetFileName).Output()
	println(string(out))
	if err != nil {
		println("failed to split video: " + err.Error() + ", file: " + filePath)
		return errors.New("failed to split video: " + filePath)
	}
	return nil
}

func (server *VideoConverterServer) tokenIsInvalid(token string) bool {
	if tokenCreationTime, ok := (*server.ActiveTokens)[token]; ok {
		if time.Since(*tokenCreationTime.CreationTime).Seconds() < tokenTimeOutSeconds {
			return false
		}
	}
	return true
}

func (serv *VideoConverterServer) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverter_DownloadServer) error {
	if serv.tokenIsInvalid(request.Id) {
		return errors.New("token is invalid or has timed out: " + request.Id)
	}
	//TODO set chunksize in a global way
	chunksize := 1000

	//TODO check if id is valid
	id := request.Id

	//TODO load corresponding file from directory
	file, err := os.Open(constants.FileDirectory + id)
	if err != nil {
		log.Fatalf("Download, Open failed: %v", err)
	}

	buf := make([]byte, chunksize)

	for {
		n, err := file.Read(buf)

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Download, Read: %v", err)
		}

		stream.Send(&videoconverter.Chunk{
			RequestType: &videoconverter.Chunk_Content{Content: buf[:n]},
		})
	}

	return nil
}

func (serv *VideoConverterServer) Delete(ctx context.Context, in *videoconverter.DeleteRequest) (*videoconverter.DeleteResponse, error) {
	if serv.tokenIsInvalid(in.Id) {
		return nil, errors.New("token is invalid or has timed out: " + in.Id)
	}
	filePath := constants.FileDirectory + in.Id + ".mp4"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("video to delete does not exist")
	}
	err := os.Remove(filePath)
	if err != nil {
		print(err.Error())
	}
	filePath = constants.FileDirectory + in.Id
	_, err = os.Stat(filePath)
	if err == nil {
		err = os.Remove(filePath)
		if err != nil {
			print(err.Error())
		}
	}
	return &videoconverter.DeleteResponse{}, nil
}

func (serv *VideoConverterServer) ConversionStatus(ctx context.Context, in *videoconverter.ConversionStatusRequest) (*videoconverter.ConversionStatusResponse, error) {
	if *(*serv.ActiveTokens)[in.StatusId].ConversionDone {
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Done}, nil
	}
	if *(*serv.ActiveTokens)[in.StatusId].ConversionStarted {
		if *(*serv.ActiveTokens)[in.StatusId].ConversionFailed {
			return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Failed}, nil
		}
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_InProgress}, nil
	}
	return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_NotStarted}, nil
}
func (serv *VideoConverterServer) resetConversionStatus(token string) {
	*(*serv.ActiveTokens)[token].ConversionStarted = false
	*(*serv.ActiveTokens)[token].ConversionFailed = false
	*(*serv.ActiveTokens)[token].ConversionDone = false
}

func GenerateRandomString() string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < tokenLength; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String() // E.g. "ExcbsVQs"
}

func (serv *VideoConverterServer) DeleteTimedOutVideosLoop() {
	for {
		for token, _ := range *serv.ActiveTokens {
			if serv.tokenIsInvalid(token) {
				filePath := constants.FileDirectory + token + ".mp4"
				_, err := os.Stat(filePath)
				if err == nil {
					println("deleting " + filePath)
					err := os.Remove(filePath)
					if err != nil {
						println(err.Error())
					}
				}
				filePath = constants.FileDirectory + token
				_, err = os.Stat(filePath)
				if err == nil {
					println("deleting " + filePath)
					err := os.Remove(filePath)
					if err != nil {
						println(err.Error())
					}
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
}
