package main

import (
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/lifeGuard"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	println(len(os.Args))
	if len(os.Args) != 5 {
		println(errors.New("invalid command line arguments, {thisIp} {port} {api-gateway ip:port} {nextLifeGuard ip:port}").Error())
		return
	}

	//TODO is the ip needed
	//ip := os.Args[1]
	port := os.Args[2]
	port = ":" + port
	//TODO create api gateway functionality for lifeguards
	//gateWayAddress := os.Args[3]
	println("running on port: " + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	lifeGuardServer := lifeGuardInterface.CreateNewLifeGuardServer()
	videoconverter.RegisterLifeGuardServer(s, &lifeGuardServer)

	targetAddress := os.Args[4]

	go func() {
		lifeGuardServer.HandleLifeGuardDuties(targetAddress)
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}