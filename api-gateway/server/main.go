package main

import (
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway"
	api_gateway2 "github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		println(errors.New("invalid command line arguments, use ./worker {port}").Error())
		return
	}
	port := os.Args[1]
	port = ":" + port
	println("running on port" + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	apiServer := api_gateway.CreateNewServer()
	api_gateway2.RegisterAPIGateWayServer(s, &apiServer)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
