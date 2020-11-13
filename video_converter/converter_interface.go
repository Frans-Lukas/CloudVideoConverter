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
	"strings"
	"time"
)

const tokenLength = 20
const tokenTimeOutSeconds = 60 * 2

type Server struct {
	videoconverter.UnimplementedVideoConverterServer
	ActiveTokens map[string]time.Time
}

func (serv *Server) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	serv.ActiveTokens[tokenString] = time.Now()
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func saveImage(fileName string, imageBytes *bytes.Buffer) error {
	imagePath := fileName
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}
	_, err = imageBytes.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write image to file: %w", err)
	}
	return nil
}

func (serv *Server) Upload(stream videoconverter.VideoConverter_UploadServer) error {

	imageData := bytes.Buffer{}
	tokenString := ""

	for {
		streamData, err := stream.Recv()
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

		if err != nil {
			if err == io.EOF {
				break
			}

			err = errors.New("failed unexpectadely while reading chunks from stream")
			return err
		}
	}

	err := saveImage("filename", &imageData)
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
func (server *Server) tokenIsInvalid(token string) bool {
	if tokenCreationTime, ok := server.ActiveTokens[token]; ok {
		if time.Since(tokenCreationTime).Seconds() < tokenTimeOutSeconds {
			return false
		}
	}
	return true
}

func (*Server) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverter_DownloadServer) error {
	//TODO set chunksize in a global way
	chunksize := 1000

	//TODO check if id is valid

	//TODO load corresponding file from directory
	file, err := os.Open("testFile")
	if err != nil {
		log.Fatalf("Download, Open: %v", err)
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

func (*Server) Delete(ctx context.Context, in *videoconverter.DeleteRequest) (*videoconverter.DeleteResponse, error) {
	return nil, nil
}

func (*Server) ConversionStatus(ctx context.Context, in *videoconverter.ConversionStatusRequest) (*videoconverter.ConversionStatusResponse, error) {
	return nil, nil
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
