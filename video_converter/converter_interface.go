package video_converter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Frans-Lukas/cloudvideoconverter/generated"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const tokenLength = 20
const tokenTimeOutSeconds = 60 * 2

type VideoConverterServer struct {
	videoconverter.UnimplementedVideoConverterServer
	ActiveTokens *map[string]time.Time
}

func CreateNewServer() VideoConverterServer {
	activeTokens := make(map[string]time.Time, 0)
	val := VideoConverterServer{
		ActiveTokens: &activeTokens,
	}
	return val
}

func (serv *VideoConverterServer) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	(*serv.ActiveTokens)[tokenString] = time.Now()
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func saveImage(fileName string, imageBytes *bytes.Buffer) error {
	imagePath := "localStorage/" + fileName + ".mp4"
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
	filePath := "localStorage/" + in.Token + ".mp4"
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
	arg2 := "localStorage/" + in.Token + "." + in.OutputType

	go func() {
		cmd := exec.Command(app, arg0, arg1, arg2)
		stdout, err := cmd.Output()
		cmd.Run()
		if err != nil {
			println("error: " + err.Error())
		}
		println("converted: ", stdout)

		file, err := os.Open(arg2)
		defer file.Close()
		if err != nil {
			println(err.Error())
			return
		}
		os.Rename(arg2, "localStorage/"+in.Token)
	}()

	return &videoconverter.ConversionResponse{}, nil
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
				return errors.New("invalid token, either timed out or nonexistant")
			}
			tokenString = token
		}
	}

	err := saveImage(tokenString, &imageData)
	if err != nil {
		return err
	}

	// once the transmission finished, send the
	// confirmation if nothign went wrong
	err = stream.SendAndClose(&videoconverter.UploadStatus{
		RetrievalToken: tokenString,
	})
	// ...

	return nil
}
func (server *VideoConverterServer) tokenIsInvalid(token string) bool {
	if tokenCreationTime, ok := (*server.ActiveTokens)[token]; ok {
		if time.Since(tokenCreationTime).Seconds() < tokenTimeOutSeconds {
			return false
		}
	}
	return true
}

func (*VideoConverterServer) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverter_DownloadServer) error {
	//TODO set chunksize in a global way
	chunksize := 1000

	//TODO check if id is valid
	id := request.Id

	//TODO load corresponding file from directory
	file, err := os.Open("localStorage/" + id)
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

func (*VideoConverterServer) Delete(ctx context.Context, in *videoconverter.DeleteRequest) (*videoconverter.DeleteResponse, error) {
	return nil, nil
}

func (*VideoConverterServer) ConversionStatus(ctx context.Context, in *videoconverter.ConversionStatusRequest) (*videoconverter.ConversionStatusResponse, error) {
	filePath := "localStorage/" + in.StatusId
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_InProgress}, nil
	}
	return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Done}, nil
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
