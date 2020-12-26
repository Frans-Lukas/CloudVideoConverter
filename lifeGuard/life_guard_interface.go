package lifeGuardInterface

import (
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type LifeGuardServer struct {
	videoconverter.UnimplementedLifeGuardServer
}

func CreateNewLifeGuardServer() LifeGuardServer {
	val := LifeGuardServer{}
	return val
}

func (server *LifeGuardServer) HandleLifeGuardDuties(targetLifeGuard string) {
	lifeGuardConnection := server.ConnectToLifeGuard(targetLifeGuard)

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 10)
		_, err := lifeGuardConnection.IsAlive(ctx, &videoconverter.IsAliveRequest{})

		if err != nil {
			log.Fatalf("response: %v", err)
		}
		println("responded!")
	}
}

func (server *LifeGuardServer) ConnectToLifeGuard(targetLifeGuard string) videoconverter.LifeGuardClient {
	println("trying to connect to: ", targetLifeGuard)

	conn, err := grpc.Dial(targetLifeGuard, grpc.WithInsecure(), grpc.WithBlock())
	for {
		//TODO if it cannot connect for a while, start it yourself
		if err == nil {
			break
		}
		conn, err = grpc.Dial(targetLifeGuard, grpc.WithInsecure(), grpc.WithBlock())
	}

	println("connected to: ", targetLifeGuard)
	lifeGuardConnection := videoconverter.NewLifeGuardClient(conn)
	return lifeGuardConnection
}

func (serv *LifeGuardServer) IsAlive(ctx context.Context, in *videoconverter.IsAliveRequest) (*videoconverter.IsAliveResponse, error) {
	return &videoconverter.IsAliveResponse{}, nil
}