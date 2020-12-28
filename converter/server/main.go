package main

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/converter"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	port = ":50051"
)

func main() {
	println(len(os.Args))
	if len(os.Args) != 4 {
		println(errors.New("invalid command line arguments, use ./worker {thisIp} {port} {api-gateway ip:port}").Error())
		return
	}

	ip := os.Args[1]
	port := os.Args[2]
	port = ":" + port
	gateWayAddress := os.Args[3]
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

	println("trying to connect to API Gateway: ", gateWayAddress)
	conn, err := grpc.Dial(gateWayAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := api_gateway.NewAPIGateWayClient(conn)
	PostServicePoint(ip, os.Args[2], c)
	conn.Close()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func PostServicePoint(Ip string, Port string, c api_gateway.APIGateWayClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	port2, _ := strconv.Atoi(Port)
	_, err := c.AddServiceEndpoint(
		ctx, &api_gateway.ServiceEndPoint{
			Ip:   Ip,
			Port: int32(port2),
		},
	)
	if err != nil {
		print(err.Error())
	}
}
