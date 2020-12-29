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
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer"
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

// server is used to implement helloworld.GreeterServer.

func main() {

	if len(os.Args) != 3 {
		println(errors.New("invalid command line arguments, use ./loadBalancer {port} {api-gateway ip:port}").Error())
		return
	}
	port := os.Args[1]
	port = ":" + port
	apiGatewayAddress := os.Args[2]

	println("running on port: " + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	videoServer := video_converter.CreateNewServer()
	videoconverter.RegisterVideoConverterLoadBalancerServer(s, &videoServer)

	//1. Load active services
	videoServer.UpdateActiveServices(apiGatewayAddress)
	videoServer.SetApiGatewayAddress(apiGatewayAddress)

	//2. Load work from database
	println("loading queue from DB")
	videoServer.LoadQueueFromDB()
	//3. Send work to services loop
	println("starting worker loop")
	go func() {
		//videoServer.IncreaseNumberOfServices()
		videoServer.WorkManagementLoop()
	}()

	println("starting server")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
