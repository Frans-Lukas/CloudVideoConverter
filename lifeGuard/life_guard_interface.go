package lifeGuardInterface

import (
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"strings"
	"time"
)

type LifeGuardServer struct {
	videoconverter.UnimplementedLifeGuardServer
	targetLifeGuard              string
	targetLifeGuardConnection    *videoconverter.LifeGuardClient
	APIGateway                   *api_gateway.APIGateWayClient
	shouldSendElectionMessage    bool
	canSendElectionMessage       bool
	newElectionRequest           *videoconverter.ElectionRequest
	id                           int32
	shouldSendCoordinatorMessage bool
	newCoordinatorRequest        *videoconverter.CoordinatorRequest
	startedElection              bool
	startedCoordination          bool
	isCoordinator                bool
	shouldRecreateRing           bool
	recreateRingSenderId         int
	yourAddress                  string
	targetIsCoordinator          bool
}

func CreateNewLifeGuardServer() LifeGuardServer {
	val := LifeGuardServer{
		targetLifeGuard:              "NOT SET",
		targetLifeGuardConnection:    nil,
		APIGateway:                   nil,
		shouldSendElectionMessage:    false,
		canSendElectionMessage:       true,
		newElectionRequest:           nil,
		id:                           -1,
		shouldSendCoordinatorMessage: false,
		startedElection:              false,
		startedCoordination:          false,
		isCoordinator:                false,
		shouldRecreateRing:           false,
		recreateRingSenderId:         -1,
		yourAddress:                  "",
		targetIsCoordinator:          false,
	}
	return val
}

func (server *LifeGuardServer) HandleLifeGuardDuties() {

	for {
		if server.shouldRecreateRing {
			server.recreateRingProcedure()
		}

		if server.targetLifeGuard == "NOT SET" || server.targetLifeGuard == "" {
			server.ConnectToLifeGuard()
			continue
		}

		server.checkIfNextLifeGuardIsAlive()

		if server.checkIfElectionShouldBeStarted() {
			server.startElection()
		}

		if server.shouldSendElectionMessage && server.canSendElectionMessage {
			server.sendElectionMessage()
		}

		if server.shouldSendCoordinatorMessage {
			server.sendCoordinatorMessage()
		}

		if server.isCoordinator {
			//TODO coordinator stuff
		}
	}
}

func (server *LifeGuardServer) ConnectToLifeGuard() {
	server.getNextLifeGuard()
	server.getLifeGuardCoordinator()

	println("trying to connect to: ", server.targetLifeGuard)

	conn, err := grpc.Dial(server.targetLifeGuard, grpc.WithInsecure(), grpc.WithBlock())
	for {
		//TODO if it cannot connect for a while, start it yourself
		if err == nil {
			break
		}
		conn, err = grpc.Dial(server.targetLifeGuard, grpc.WithInsecure(), grpc.WithBlock())
	}

	println("connected to: ", server.targetLifeGuard)
	lifeGuardConnection := videoconverter.NewLifeGuardClient(conn)
	server.targetLifeGuardConnection = &lifeGuardConnection
}

func (server *LifeGuardServer) IsAlive(ctx context.Context, in *videoconverter.IsAliveRequest) (*videoconverter.IsAliveResponse, error) {
	return &videoconverter.IsAliveResponse{}, nil
}

func (server *LifeGuardServer) Election(ctx context.Context, in *videoconverter.ElectionRequest) (*videoconverter.ElectionResponse, error) {
	if server.startedElection {
		server.shouldSendCoordinatorMessage = true
		server.newElectionRequest = in
	} else if in.HighestProcessNumber == server.id {
		server.shouldSendElectionMessage = true
		server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id}
	} else if in.HighestProcessNumber < server.id {
		server.shouldSendElectionMessage = true
		server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id}
	} else {
		server.shouldSendElectionMessage = true
		server.newElectionRequest = in
	}
	return &videoconverter.ElectionResponse{}, nil
}

func (server *LifeGuardServer) Coordinator(ctx context.Context, in *videoconverter.CoordinatorRequest) (*videoconverter.CoordinatorResponse, error) {
	if in.HighestProcessNumber == server.id {
		server.isCoordinator = true
	}

	if server.startedCoordination {
		server.newElectionRequest = nil
	} else {
		server.shouldSendCoordinatorMessage = true
		server.newCoordinatorRequest = in
	}

	return &videoconverter.CoordinatorResponse{}, nil
}

func (server *LifeGuardServer) RecreateRing(ctx context.Context, in *videoconverter.RecreateRingRequest) (*videoconverter.RecreateRingResponse, error) {
	if in.InitialSenderId == server.id {
		if server.shouldSendElectionMessage {
			server.canSendElectionMessage = true
			server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id}
		}
	} else {
		server.shouldRecreateRing = true
	}
	return &videoconverter.RecreateRingResponse{}, nil
}

func (server *LifeGuardServer) checkIfNextLifeGuardIsAlive() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).IsAlive(ctx, &videoconverter.IsAliveRequest{})

	if err != nil {
		log.Fatalf("response to IsAlive: %v", err)
		if server.targetIsCoordinator {
			server.shouldSendElectionMessage = true
			server.canSendElectionMessage = false
		}
		server.removeTargetLifeGuard()
		server.shouldRecreateRing = true
		server.recreateRingSenderId = int(server.id)
	}
}

func (server *LifeGuardServer) startElection() {
	server.startedElection = true
	server.shouldSendElectionMessage = true
	server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id}
}

func (server *LifeGuardServer) sendElectionMessage() {
	println("Sending ElectionMessage")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).Election(ctx, server.newElectionRequest)

	if err != nil {
		log.Fatalf("response to election message: %v", err)
	}
	println("responded!")
	server.newElectionRequest = nil
	server.shouldSendElectionMessage = false
}

func (server *LifeGuardServer) sendCoordinatorMessage() {
	println("Sending CoordinatorMessage")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).Coordinator(ctx, server.newCoordinatorRequest)

	if err != nil {
		log.Fatalf("response to coordinator message: %v", err)
	}
	println("responded!")
	server.newCoordinatorRequest = nil
	server.shouldSendCoordinatorMessage = false
}

func (server *LifeGuardServer) SetupAPIConnections(Ip string, Port string, c api_gateway.APIGateWayClient) {
	intPort, err := strconv.Atoi(Port)
	if err != nil {
		log.Fatalf("Port not int: " + err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	res, err := c.AddLifeGuardNode(ctx, &api_gateway.AddLifeGuardNodeRequest{Ip:Ip, Port:int32(intPort)})
	if err != nil {
		log.Fatalf("AddLifeGuardNode: " + err.Error())
	}
	server.id = res.NewLifeGuardId

	server.APIGateway = &c

	server.shouldRecreateRing = true
	server.recreateRingSenderId = int(server.id)

	server.yourAddress = Ip + ":" + Port
}

func (server *LifeGuardServer) checkIfElectionShouldBeStarted() bool {
	//TODO figure out how to implement this
	return false
}

func (server *LifeGuardServer) recreateRingProcedure() {
	server.getNextLifeGuard()
	if server.targetLifeGuard == "NOT SET" || server.targetLifeGuard == "" {
		println("recreateRingProcedure: target not set")
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).RecreateRing(ctx, &videoconverter.RecreateRingRequest{InitialSenderId:int32(server.recreateRingSenderId)})

	if err != nil {
		println("recreateRingProcedure: " + err.Error())
		return
	}

	println("Sent recreateRing message")
	server.shouldRecreateRing = false
}

func (server *LifeGuardServer) getNextLifeGuard() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	res, err := (*server.APIGateway).GetNextLifeGuard(ctx, &api_gateway.GetNextLifeGuardRequest{LifeGuardId: server.id})

	if err != nil {
		println("getNextLifeGuard: " + err.Error())
		server.targetLifeGuard = "NOT SET"
		return
	}

	println("Next LifeGuard is: " + res.Ip + ":" + strconv.Itoa(int(res.Port)))
	server.targetLifeGuard = res.Ip + ":" + strconv.Itoa(int(res.Port))
}

func (server *LifeGuardServer) getLifeGuardCoordinator() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	res, err := (*server.APIGateway).GetLifeGuardCoordinator(ctx, &api_gateway.GetLifeGuardCoordinatorRequest{})

	if err != nil {
		println("getLifeGuardCoordinator: " + err.Error())
		server.isCoordinator = false
		return
	}

	coordinatorLifeGuard := res.Ip + ":" + strconv.Itoa(int(res.Port))

	if coordinatorLifeGuard == server.yourAddress {
		server.isCoordinator = true
		server.targetIsCoordinator = false
	} else if coordinatorLifeGuard == server.targetLifeGuard {
		server.isCoordinator = false
		server.targetIsCoordinator = true
	} else {
		server.isCoordinator = false
		server.targetIsCoordinator = false
	}
}

func (server *LifeGuardServer) removeTargetLifeGuard() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	splitAddress := strings.Split(server.targetLifeGuard, ":")
	port, err := strconv.Atoi(splitAddress[1])

	if err != nil {
		log.Fatalf("removeTargetLifeGuard: invalid address string")
	}

	_, err = (*server.APIGateway).RemoveLifeGuardNode(ctx, &api_gateway.RemoveLifeGuardNodeRequest{Ip:splitAddress[0], Port:int32(port)})

	if err != nil {
		println("removeTargetLifeGuard: " + err.Error())
	}

	server.targetLifeGuard = "NOT SET"
}
