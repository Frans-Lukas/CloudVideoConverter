package video_converter

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/server/items"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const tokenLength = 20

type Server struct {
	videoconverter.UnimplementedVideoConverterServer
	ActiveTokens map[items.Token]bool
}

func (serv *Server) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	token := items.Token{CreationTime: time.Now(), TokenString: tokenString}
	serv.ActiveTokens[token] = true
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func (*Server) Upload(stream videoconverter.VideoConverter_UploadServer) error {
	return nil
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

	EOFError := errors.New("EOF")

	writing := true
	for writing {
		n, err := file.Read(buf)

		if err != nil && errors.Is(err, EOFError) {
			writing = false
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
