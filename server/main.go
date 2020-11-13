/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/server/items"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

const (
	port        = ":50051"
	tokenLength = 20
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	videoconverter.UnimplementedVideoConverterServer
	ActiveTokens map[items.Token]bool
}

func (serv *server) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	token := items.Token{CreationTime: time.Now(), TokenString: tokenString}
	serv.ActiveTokens[token] = true
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func (*server) Upload(stream videoconverter.VideoConverter_UploadServer) error {
	return nil
}

func (*server) ConversionStatus(ctx context.Context, in *videoconverter.ConversionStatusRequest) (*videoconverter.ConversionStatusResponse, error) {
	return nil, nil
}

func (*server) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverter_DownloadServer) error {
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

func (*server) Delete(ctx context.Context, in *videoconverter.DeleteRequest) (*videoconverter.DeleteResponse, error) {
	return nil, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	videoconverter.RegisterVideoConverterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
