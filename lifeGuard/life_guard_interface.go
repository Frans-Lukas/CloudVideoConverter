package lifeGuardInterface

import (
	"bytes"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"math"
	"os/exec"
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
	shouldStartElection          bool
	newElectionRequest           *videoconverter.ElectionRequest
	id                           int32
	shouldSendCoordinatorMessage bool
	newCoordinatorRequest        *videoconverter.CoordinatorRequest
	startedElection              bool
	startedCoordination          bool
	isCoordinator                bool
	isCoordinatorOutput          chan<- *bool
	shouldRecreateRing           bool
	recreateRingSenderId         int
	yourAddress                  string
	targetIsCoordinator          bool
	shouldSetCoordinator         bool
	shouldRestartDeadLifeGuards  bool
	loadBalancerPort             int
}

func CreateNewLifeGuardServer(coordinatorStatus chan *bool, loadBalancerPort int) LifeGuardServer {
	val := LifeGuardServer{
		targetLifeGuard:              "NOT SET",
		targetLifeGuardConnection:    nil,
		APIGateway:                   nil,
		shouldSendElectionMessage:    false,
		shouldStartElection:          false,
		newElectionRequest:           nil,
		id:                           -1,
		shouldSendCoordinatorMessage: false,
		startedElection:              false,
		startedCoordination:          false,
		isCoordinator:                false,
		isCoordinatorOutput:          coordinatorStatus,
		shouldRecreateRing:           false,
		recreateRingSenderId:         -1,
		yourAddress:                  "",
		targetIsCoordinator:          false,

		shouldSetCoordinator:        false,
		shouldRestartDeadLifeGuards: false,

		loadBalancerPort: loadBalancerPort,
	}
	return val
}

func (server *LifeGuardServer) HandleLifeGuardDuties() {

	i := 0.0

	for {
		time.Sleep(time.Second * 3)

		if server.shouldRestartDeadLifeGuards && math.Mod(i, 4) == 0 {
			server.restartDeadLifeGuards()
		}

		server.checkIfNextLifeGuardIsAlive()

		if server.shouldRecreateRing {
			server.recreateRingProcedure()
		}

		if server.targetLifeGuard == "NOT SET" || server.targetLifeGuard == "" {
			server.ConnectToLifeGuard()
			continue
		}

		if server.shouldSendElectionMessage {
			server.sendElectionMessage()
		}

		if server.shouldSendCoordinatorMessage {
			server.sendCoordinatorMessage()
		}

		if server.shouldSetCoordinator {
			server.setCoordinator()
		}

		if server.shouldStartElection {
			server.startElection()
		}

		i++
	}
}

func (server *LifeGuardServer) ConnectToLifeGuard() {
	server.getNextLifeGuard()
	server.getLifeGuardCoordinator()

	println("trying to connect to: ", server.targetLifeGuard)

	conn, err := grpc.Dial(server.targetLifeGuard, grpc.WithInsecure(), grpc.WithTimeout(time.Second*10))
	for i := 0; i < 3; i++ {
		if err == nil {
			break
		}
		println("trying to connect to: ", server.targetLifeGuard)
		conn, err = grpc.Dial(server.targetLifeGuard, grpc.WithInsecure(), grpc.WithTimeout(time.Second*10))
	}

	println("connected to: ", server.targetLifeGuard)
	lifeGuardConnection := videoconverter.NewLifeGuardClient(conn)
	server.targetLifeGuardConnection = &lifeGuardConnection
}

func (server *LifeGuardServer) IsAlive(ctx context.Context, in *videoconverter.IsAliveRequest) (*videoconverter.IsAliveResponse, error) {
	return &videoconverter.IsAliveResponse{}, nil
}

func (server *LifeGuardServer) Election(ctx context.Context, in *videoconverter.ElectionRequest) (*videoconverter.ElectionResponse, error) {
	println("Received election message with ID: " + strconv.Itoa(int(in.HighestProcessNumber)) + " initially from: " + strconv.Itoa(int(in.HighestProcessNumber)))

	if in.InitialSenderId == server.id {
		server.shouldSendCoordinatorMessage = true
		server.startedCoordination = true
		server.shouldSendElectionMessage = false
		server.startedElection = false
		server.newCoordinatorRequest = &videoconverter.CoordinatorRequest{HighestProcessNumber: in.HighestProcessNumber, InitialSenderId: server.id}
	} else if in.HighestProcessNumber <= server.id {
		server.shouldSendElectionMessage = true
		server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id, InitialSenderId: in.InitialSenderId}
	} else {
		server.shouldSendElectionMessage = true
		server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: in.HighestProcessNumber, InitialSenderId: in.InitialSenderId}
	}
	return &videoconverter.ElectionResponse{}, nil
}

func (server *LifeGuardServer) Coordinator(ctx context.Context, in *videoconverter.CoordinatorRequest) (*videoconverter.CoordinatorResponse, error) {
	if in.HighestProcessNumber == server.id {
		println("should set is coordinator")
		//server.updateIsCoordinator(true)
		server.shouldSetCoordinator = true
	}

	if in.InitialSenderId == server.id {
		server.newElectionRequest = nil
		server.startedCoordination = false
	} else {
		server.shouldSendCoordinatorMessage = true
		server.newCoordinatorRequest = in
	}

	return &videoconverter.CoordinatorResponse{}, nil
}

func (server *LifeGuardServer) RecreateRing(ctx context.Context, in *videoconverter.RecreateRingRequest) (*videoconverter.RecreateRingResponse, error) {
	if in.InitialSenderId == server.id {
		println("RecreateRing Completed!!!")
		if server.shouldSendElectionMessage {
			server.shouldStartElection = true
			server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id, InitialSenderId: server.id}
		}
	} else {
		println("recreateRing id: " + strconv.Itoa(int(in.InitialSenderId)))
		server.shouldRecreateRing = true
		server.recreateRingSenderId = int(in.InitialSenderId)
	}
	return &videoconverter.RecreateRingResponse{}, nil
}

func (server *LifeGuardServer) checkIfNextLifeGuardIsAlive() {

	if server.targetLifeGuard == "NOT SET" || server.targetLifeGuard == "" {
		println("checkIfNextLifeGuardIsAlive: target not set")
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).IsAlive(ctx, &videoconverter.IsAliveRequest{})

	if err != nil {
		println("response to IsAlive: %v", err)
		if server.targetIsCoordinator {
			server.shouldRestartDeadLifeGuards = true
			server.shouldStartElection = true
		}
		server.removeTargetLifeGuard()
		server.shouldRecreateRing = true
		server.recreateRingSenderId = int(server.id)
	}
}

func (server *LifeGuardServer) startElection() {
	println("starting election!")
	server.startedElection = true
	server.shouldSendElectionMessage = true
	server.shouldStartElection = false
	server.newElectionRequest = &videoconverter.ElectionRequest{HighestProcessNumber: server.id, InitialSenderId: server.id}
}

func (server *LifeGuardServer) sendElectionMessage() {
	println("Sending ElectionMessage with id: " + strconv.Itoa(int(server.newElectionRequest.HighestProcessNumber)))
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).Election(ctx, server.newElectionRequest)

	if err != nil {
		println("response to election message: %v", err)
		server.targetLifeGuard = "NOT SET"
		return
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
		println("response to coordinator message: " + err.Error())
		return
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
	res, err := c.AddLifeGuardNode(ctx, &api_gateway.AddLifeGuardNodeRequest{Ip: Ip, Port: int32(intPort)})
	if err != nil {
		log.Fatalf("AddLifeGuardNode: " + err.Error())
	}
	server.id = res.NewLifeGuardId

	server.APIGateway = &c

	server.shouldRecreateRing = true
	server.recreateRingSenderId = int(server.id)

	server.yourAddress = Ip + ":" + Port
}

func (server *LifeGuardServer) recreateRingProcedure() {
	server.ConnectToLifeGuard()
	if server.targetLifeGuard == "NOT SET" || server.targetLifeGuard == "" {
		println("recreateRingProcedure: target not set")
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.targetLifeGuardConnection).RecreateRing(ctx, &videoconverter.RecreateRingRequest{InitialSenderId: int32(server.recreateRingSenderId)})

	if err != nil {
		println("recreateRingProcedure: " + err.Error())
		server.targetLifeGuard = "NOT SET"
		return
	}

	println("Sent recreateRing message")
	server.shouldRecreateRing = false
}

func (server *LifeGuardServer) getNextLifeGuard() {
	println("Seeking next lifeGuard for id: ", server.id)
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
		server.updateIsCoordinator(false)
		return
	}

	if res.Port == -1 || res.Ip == "" {
		println("Coordinator not set")
		server.startElection()
		return
	}

	coordinatorLifeGuard := res.Ip + ":" + strconv.Itoa(int(res.Port))

	println("Coordinator is: " + coordinatorLifeGuard)

	if coordinatorLifeGuard == server.yourAddress {
		server.updateIsCoordinator(true)
		server.targetIsCoordinator = false
	} else if coordinatorLifeGuard == server.targetLifeGuard {
		server.updateIsCoordinator(false)
		server.targetIsCoordinator = true
	} else {
		server.updateIsCoordinator(false)
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

	_, err = (*server.APIGateway).RemoveLifeGuardNode(ctx, &api_gateway.RemoveLifeGuardNodeRequest{Ip: splitAddress[0], Port: int32(port)})

	if err != nil {
		println("removeTargetLifeGuard: " + err.Error())
	}

	server.targetLifeGuard = "NOT SET"
}

func (server *LifeGuardServer) setCoordinator() {
	println("Set coordinator to me")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	_, err := (*server.APIGateway).SetLifeGuardCoordinator(ctx, &api_gateway.SetLifeGuardCoordinatorRequest{LifeGuardId: server.id, LoadBalancerPort: int32(server.loadBalancerPort)})

	if err != nil {
		println("setCoordinator: " + err.Error())
		server.shouldSetCoordinator = false
		return
	}

	server.updateIsCoordinator(true)
}

func (server *LifeGuardServer) updateIsCoordinator(b bool) {
	if b {
		println("is coordinator")
	} else {
		println("is NOT coordinator")
	}
	server.isCoordinator = b
	server.isCoordinatorOutput <- &b
}

func (server *LifeGuardServer) getDesiredLifeGuards() *api_gateway.GetMaxLifeGuardsResponse {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	res, err := (*server.APIGateway).GetMaxLifeGuards(ctx, &api_gateway.GetMaxLifeGuardsRequest{})
	if err != nil {
		println("getDesiredLifeGuards: " + err.Error())
		return res
	}

	return res
}

func (server *LifeGuardServer) restartDeadLifeGuards() {
	desiredNumber := server.getDesiredLifeGuards()

	if desiredNumber.IsCurrentlyMaxNumber {
		println("Is max lifeGuards!")
		server.shouldRestartDeadLifeGuards = false
		return
	}

	go server.executeLifeGuardRestart(int(desiredNumber.MaxLifeGuards))
}

func (server *LifeGuardServer) executeLifeGuardRestart(desiredNumber int) {
	println("./scripts/tfScripts/LoadBalancerWithoutAPIGateWay/startLoadBalancerVMsFromLifeGuard.sh", desiredNumber)
	cmd := exec.Command("./scripts/tfScripts/LoadBalancerWithoutAPIGateWay/startLoadBalancerVMsFromLifeGuard.sh", strconv.Itoa(desiredNumber))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		println("executeLifeGuardRestart: " + err.Error())
		println(out.String())
		println(stderr.String())

		return
	}

	println("done with executeLifeGuardRestart")
}
