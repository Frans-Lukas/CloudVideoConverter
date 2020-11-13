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
	"github.com/Frans-Lukas/cloudvideoconverter/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/server/items"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
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
	tokenString := generateRandomString()
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

func generateRandomString() string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < tokenLength; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String() // E.g. "ExcbsVQs"
}
