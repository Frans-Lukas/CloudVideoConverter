package main

import (
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/converter"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	port = ":50051"
)

func main() {
	if len(os.Args) != 2 {
		//TODO update so that loadbalancer and converter does not have the same input
		println(errors.New("invalid command line arguments, use ./worker {port}").Error())
		return
	}

	port := os.Args[1]
	port = ":" + port
	println("running on port: " + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	videoConverterServer := converter.CreateNewVideoConverterServiceServer()
	videoconverter.RegisterVideoConverterServiceServer(s, &videoConverterServer)

	go func() {
		videoConverterServer.HandleConversionsLoop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}