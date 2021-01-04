package lifeGuardInterface

import (
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type LifeGuardServer struct {
	videoconverter.UnimplementedLifeGuardServer
	targetLifeGuard              string
	shouldSendElectionMessage    bool
	newElectionRequest           *videoconverter.ElectionRequest
	id                           int32
	shouldSendCoordinatorMessage bool
	newCoordinatorRequest        *videoconverter.CoordinatorRequest
	startedElection              bool
	startedCoordination          bool
	isCoordinator                bool
}

func CreateNewLifeGuardServer() LifeGuardServer {
	val := LifeGuardServer{
		targetLifeGuard:              "NOT SET",
		shouldSendElectionMessage:    false,
		newElectionRequest:           nil,
		id:                           -1,
		shouldSendCoordinatorMessage: false,
		startedElection:              false,
		startedCoordination:          false,
		isCoordinator:                false,
	}
	return val
}

func (server *LifeGuardServer) HandleLifeGuardDuties() {
	lifeGuardConnection := server.ConnectToLifeGuard()

	for {
		server.checkIfNextLifeGuardIsAlive(lifeGuardConnection)

		if server.checkIfElectionShouldBeStarted() {
			server.startElection()
		}

		if server.shouldSendElectionMessage {
			server.sendElectionMessage(lifeGuardConnection)
		}

		if server.shouldSendCoordinatorMessage {
			server.sendCoordinatorMessage(lifeGuardConnection)
		}

		if server.isCoordinator {
			//TODO coordinator stuff
		}
	}
}

func (server *LifeGuardServer) ConnectToLifeGuard() videoconverter.LifeGuardClient {
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
	return lifeGuardConnection
}

func (serv *LifeGuardServer) IsAlive(ctx context.Context, in *videoconverter.IsAliveRequest) (*videoconverter.IsAliveResponse, error) {
	return &videoconverter.IsAliveResponse{}, nil
}

func (serv *LifeGuardServer) Election(ctx context.Context, in *videoconverter.ElectionRequest) (*videoconverter.ElectionResponse, error) {
	if serv.startedElection {
		serv.shouldSendCoordinatorMessage = true
		serv.newElectionRequest = in
	} else if in.HighestProcessNumber == serv.id {
		serv.shouldSendElectionMessage = true
		serv.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: serv.id}
	} else if in.HighestProcessNumber < serv.id {
		serv.shouldSendElectionMessage = true
		serv.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: serv.id}
	} else {
		serv.shouldSendElectionMessage = true
		serv.newElectionRequest = in
	}
	return &videoconverter.ElectionResponse{}, nil
}

func (serv *LifeGuardServer) Coordinator(ctx context.Context, in *videoconverter.CoordinatorRequest) (*videoconverter.CoordinatorResponse, error) {
	if in.HighestProcessNumber == serv.id {
		serv.isCoordinator = true
	}

	if serv.startedCoordination {
		serv.newElectionRequest = nil
	} else {
		serv.shouldSendCoordinatorMessage = true
		serv.newCoordinatorRequest = in
	}

	return &videoconverter.CoordinatorResponse{}, nil
}

func (server *LifeGuardServer) checkIfNextLifeGuardIsAlive(client videoconverter.LifeGuardClient) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := client.IsAlive(ctx, &videoconverter.IsAliveRequest{})

	if err != nil {
		log.Fatalf("response to IsAlive: %v", err)
	}
}

func (server *LifeGuardServer) startElection() {
	server.startedElection = true
	server.shouldSendElectionMessage = true
	server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber:server.id}
}

func (server *LifeGuardServer) sendElectionMessage(client videoconverter.LifeGuardClient) {
	println("Sending ElectionMessage")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := client.Election(ctx, server.newElectionRequest)

	if err != nil {
		log.Fatalf("response to election message: %v", err)
	}
	println("responded!")
	server.newElectionRequest = nil
	server.shouldSendElectionMessage = false
}

func (server *LifeGuardServer) sendCoordinatorMessage(client videoconverter.LifeGuardClient) {
	println("Sending CoordinatorMessage")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := client.Coordinator(ctx, server.newCoordinatorRequest)

	if err != nil {
		log.Fatalf("response to coordinator message: %v", err)
	}
	println("responded!")
	server.newCoordinatorRequest = nil
	server.shouldSendCoordinatorMessage = false
}

func (server *LifeGuardServer) SetupAPIConnections(Ip string, Port string, c api_gateway.APIGateWayClient) {
	//TODO get id
	//TODO post address
	//TODO get target lifeguard
}

func (server *LifeGuardServer) checkIfElectionShouldBeStarted() bool {
	//TODO figure out how to implement this
	return false
}
