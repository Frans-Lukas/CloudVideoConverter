package server

import (
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/lifeGuard"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"time"
)

func StartLifeGuard(ip string, port string, loadBalancerPort int, gateWayAddress string, coordinatorStatus chan *bool) {
	/*println(len(os.Args))
	if len(os.Args) != 4 {
		println(errors.New("invalid command line arguments, {thisIp} {port} {api-gateway ip:port}").Error())
		return
	}*/

	//ip := os.Args[1]
	//port := os.Args[2]
	portNumber := port
	port = ":" + port
	//gateWayAddress := os.Args[3]
	println("running on port: " + port)
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	lifeGuardServer := lifeGuardInterface.CreateNewLifeGuardServer(coordinatorStatus, loadBalancerPort)
	videoconverter.RegisterLifeGuardServer(s, &lifeGuardServer)

	println("trying to connect to API Gateway: ", gateWayAddress)
	conn, err := grpc.Dial(gateWayAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := api_gateway.NewAPIGateWayClient(conn)
	lifeGuardServer.SetupAPIConnections(ip, portNumber, c)

	go func() {
		lifeGuardServer.HandleLifeGuardDuties()
	}()


	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}