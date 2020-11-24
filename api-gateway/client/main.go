package main

import (
	"context"
	"errors"
	api_gateway "github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {

	if len(os.Args) != 5 {
		println(errors.New("invalid command line arguments, use ./worker {ip} {port} {serviceAddr} {servicePort}").Error())
		return
	}
	ip := os.Args[1]
	port := os.Args[2]

	serviceIp := os.Args[3]
	servicePort := os.Args[4]

	address := ip + ":" + port
	println("trying to connect to: ", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := api_gateway.NewAPIGateWayClient(conn)
	PostServicePoint(serviceIp, servicePort, c)

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
