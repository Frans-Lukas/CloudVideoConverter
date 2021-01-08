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
	if len(os.Args) != 5 {
		println(errors.New("invalid command line arguments, use ./worker {thisIp} {port} {api-gateway ip:port} {name}").Error())
		return
	}

	ip := os.Args[1]
	port := os.Args[2]
	port = ":" + port
	thisIp := ip + port
	gateWayAddress := os.Args[3]
	println("running on port: " + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	serv := grpc.NewServer()

	videoConverterServer := converter.CreateNewVideoConverterServiceServer(thisIp, os.Args[4])
	videoconverter.RegisterVideoConverterServiceServer(serv, &videoConverterServer)

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
	go func() {
		port, err := strconv.Atoi(os.Args[2])
		defer conn.Close()
		if err != nil {
			log.Fatalf("invalid port argument")
		}
		PostServicePointLoop(c, ip, int32(port))
	}()

	if err := serv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func PostServicePoint(Ip string, Port int32, c api_gateway.APIGateWayClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	println("adding service endpoint to APIGateway with port: ", Port, " and IP: ", Ip)
	_, err := c.AddServiceEndpoint(
		ctx, &api_gateway.ServiceEndPoint{
			Ip:   Ip,
			Port: Port,
		},
	)
	if err != nil {
		print(err.Error())
	}
}

func PostServicePointLoop(c api_gateway.APIGateWayClient, thisIp string, thisPort int32) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		res, err := c.GetActiveServiceEndpoints(ctx, &api_gateway.ServiceEndPointsRequest{})
		if err == nil {
			found := false
			for _, v := range (*res).EndPoint {
				if v.Ip == thisIp && v.Port == thisPort {
					found = true
				}
			}
			if !found {
				PostServicePoint(thisIp, thisPort, c)
			}
		}
		time.Sleep(time.Second * 20)
	}

}
