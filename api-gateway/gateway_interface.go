package api_gateway

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/items"
	"sort"
	"strconv"
)

type APIGatewayServer struct {
	api_gateway.APIGateWayServer
	endPoints *map[items.EndPoint]bool

	lifeGuards         *map[int]items.LifeGuard
	nextLifeGuardId    int
	currentCoordinator items.LifeGuard
}

func CreateNewServer() APIGatewayServer {
	endpoints := make(map[items.EndPoint]bool, 0)
	lifeGuards := make(map[int]items.LifeGuard, 0)
	val := APIGatewayServer{
		endPoints:          &endpoints,
		lifeGuards:         &lifeGuards,
		nextLifeGuardId:    0,
		currentCoordinator: items.LifeGuard{Ip: "", Port: -1},
	}
	return val
}

func (serv *APIGatewayServer) AddServiceEndpoint(
	ctx context.Context, in *api_gateway.ServiceEndPoint,
) (*api_gateway.AddedServiceEndPoint, error) {
	newEndPoint := items.EndPoint{
		Ip:   in.Ip,
		Port: int(in.Port),
	}
	(*serv.endPoints)[newEndPoint] = true
	println("added service: " + in.Ip + ":" + strconv.Itoa(int(in.Port)))
	return &api_gateway.AddedServiceEndPoint{}, nil
}

func (serv *APIGatewayServer) DisableServiceEndpoint(
	ctx context.Context, in *api_gateway.DisableServiceEndPointRequest,
) (*api_gateway.DisabledServiceEndPointResponse, error) {
	found := false
	for endPoint := range *serv.endPoints {
		println("iterating ip: ", endPoint.Ip, " comparing with in.ip: ", in.Ip)
		if endPoint.Ip == in.Ip {
			println("removing service from gateway, addr: ", in.Ip, ":", in.Port)
			(*serv.endPoints)[endPoint] = false
			delete(*serv.endPoints, endPoint)
			found = true
		}
	}
	if !found {
		return nil, errors.New("tried disabling endpoint but it did not exist")
	}
	return &api_gateway.DisabledServiceEndPointResponse{}, nil
}

func (serv *APIGatewayServer) GetActiveServiceEndpoints(
	ctx context.Context, in *api_gateway.ServiceEndPointsRequest,
) (*api_gateway.ServiceEndPointList, error) {
	outList := make([]*api_gateway.ServiceEndPoint, 0)
	for endPoint, isActive := range *serv.endPoints {
		if isActive {
			outList = append(
				outList, &api_gateway.ServiceEndPoint{Ip: endPoint.Ip, Port: int32(endPoint.Port)},
			)
		}
	}
	return &api_gateway.ServiceEndPointList{EndPoint: outList}, nil
}

func (serv *APIGatewayServer) AddLifeGuardNode(
	ctx context.Context, in *api_gateway.AddLifeGuardNodeRequest,
) (*api_gateway.AddLifeGuardNodeResponse, error) {
	newLifeGuard := items.LifeGuard{
		Ip:   in.Ip,
		Port: int(in.Port),
	}
	(*serv.lifeGuards)[serv.nextLifeGuardId] = newLifeGuard
	println("added lifeGuard: " + in.Ip + ":" + strconv.Itoa(int(in.Port)) + " with id: " + strconv.Itoa(serv.nextLifeGuardId))

	serv.nextLifeGuardId += 1

	return &api_gateway.AddLifeGuardNodeResponse{NewLifeGuardId: int32(serv.nextLifeGuardId - 1)}, nil
}

func (serv *APIGatewayServer) RemoveLifeGuardNode(
	ctx context.Context, in *api_gateway.RemoveLifeGuardNodeRequest,
) (*api_gateway.RemoveLifeGuardNodeResponse, error) {
	for k, v := range *serv.lifeGuards {
		if v.Port == int(in.Port) && v.Ip == in.Ip {
			println("Deleting lifeGuard: " + in.Ip + ":" + strconv.Itoa(int(in.Port)) + " with id: " + strconv.Itoa(k))
			delete(*serv.lifeGuards, k)
			break
		}
	}
	return &api_gateway.RemoveLifeGuardNodeResponse{}, nil
}

func (serv *APIGatewayServer) SetLifeGuardCoordinator(
	ctx context.Context, in *api_gateway.SetLifeGuardCoordinatorRequest,
) (*api_gateway.SetLifeGuardCoordinatorResponse, error) {
	lifeGuard, found := (*serv.lifeGuards)[int(in.LifeGuardId)]

	if !found {
		println("Got request to add lifeGuard with invalid ID.")
		return &api_gateway.SetLifeGuardCoordinatorResponse{}, errors.New("there does not exist a lifeGuard with this ID")
	}

	serv.currentCoordinator = lifeGuard
	println("New Coordinator: " + lifeGuard.Ip + ":" + strconv.Itoa(lifeGuard.Port) + " with id: " + strconv.Itoa(int(in.LifeGuardId)))
	return &api_gateway.SetLifeGuardCoordinatorResponse{}, nil
}

func (serv *APIGatewayServer) GetLifeGuardCoordinator(
	ctx context.Context, in *api_gateway.GetLifeGuardCoordinatorRequest,
) (*api_gateway.GetLifeGuardCoordinatorResponse, error) {

	// First coordinator will just be the first lifeguard that joins
	if serv.currentCoordinator.Ip == "" && serv.currentCoordinator.Port == -1 {
		if len(*serv.lifeGuards) > 0 {
			serv.currentCoordinator = (*serv.lifeGuards)[0]
		} else {
			println("Received request for coordinator before any lifeGuard has joined.")
			return &api_gateway.GetLifeGuardCoordinatorResponse{Ip: "", Port: -1}, errors.New("request for coordinator before any lifeGuard has joined")
		}
	}

	return &api_gateway.GetLifeGuardCoordinatorResponse{Ip: serv.currentCoordinator.Ip, Port: int32(serv.currentCoordinator.Port)}, nil
}

func (serv *APIGatewayServer) GetNextLifeGuard(
	ctx context.Context, in *api_gateway.GetNextLifeGuardRequest,
) (*api_gateway.GetNextLifeGuardResponse, error) {
	keys := make([]int, 0, len(*serv.lifeGuards))
	for k := range *serv.lifeGuards {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for i, k := range keys {
		if k == int(in.LifeGuardId) {
			var nextLifeGuard items.LifeGuard
			if i == len(keys)-1 {
				nextLifeGuard = (*serv.lifeGuards)[keys[0]]
			} else {
				nextLifeGuard = (*serv.lifeGuards)[keys[i+1]]
			}
			return &api_gateway.GetNextLifeGuardResponse{Ip: nextLifeGuard.Ip, Port: int32(nextLifeGuard.Port)}, nil
		}
	}

	println("Did not find nextLifeguard for lifeGuard id: " + strconv.Itoa(int(in.LifeGuardId)))
	return &api_gateway.GetNextLifeGuardResponse{Ip: "", Port: int32(-1)}, errors.New("could not find you in list of lifeGuards")
}
