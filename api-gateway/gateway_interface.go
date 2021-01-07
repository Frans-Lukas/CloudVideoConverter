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
	api_gateway.UnimplementedAPIGateWayServer
	endPoints *map[items.EndPoint]bool

	lifeGuards          *map[int]items.LifeGuard
	nextLifeGuardId     int
	currentCoordinator  items.LifeGuard
	currentLoadBalancer items.EndPoint
	maxLifeGuards       int32
}

func CreateNewServer() APIGatewayServer {
	endpoints := make(map[items.EndPoint]bool, 0)
	lifeGuards := make(map[int]items.LifeGuard, 0)
	val := APIGatewayServer{
		endPoints:           &endpoints,
		lifeGuards:          &lifeGuards,
		nextLifeGuardId:     0,
		currentCoordinator:  items.LifeGuard{Ip: "", Port: -1},
		currentLoadBalancer: items.EndPoint{Ip: "", Port: -1},
		maxLifeGuards:       0,
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

	// check if lifeGuard existed previously
	for id, lifeGuard := range *serv.lifeGuards {
		if lifeGuard.Ip == in.Ip && lifeGuard.Port == int(in.Port) {
			return &api_gateway.AddLifeGuardNodeResponse{NewLifeGuardId: int32(id)}, nil
		}
	}

	(*serv.lifeGuards)[serv.nextLifeGuardId] = newLifeGuard
	println("added lifeGuard: " + in.Ip + ":" + strconv.Itoa(int(in.Port)) + " with id: " + strconv.Itoa(serv.nextLifeGuardId))

	serv.nextLifeGuardId += 1

	if int32(len(*serv.lifeGuards)) > serv.maxLifeGuards {
		serv.maxLifeGuards = int32(len(*serv.lifeGuards))
		println("max lifeguards increased to: " + strconv.Itoa(int(serv.maxLifeGuards)))
	}

	return &api_gateway.AddLifeGuardNodeResponse{NewLifeGuardId: int32(serv.nextLifeGuardId - 1)}, nil
}

func (serv *APIGatewayServer) RemoveLifeGuardNode(
	ctx context.Context, in *api_gateway.RemoveLifeGuardNodeRequest,
) (*api_gateway.RemoveLifeGuardNodeResponse, error) {
	println("trying to delete lifeguard ", in.Ip, " port: ", in.Port)
	for k, v := range *serv.lifeGuards {
		if v.Port == int(in.Port) && v.Ip == in.Ip {
			println("Deleting lifeGuard: " + in.Ip + ":" + strconv.Itoa(int(in.Port)) + " with id: " + strconv.Itoa(k))
			delete(*serv.lifeGuards, k)

			if serv.currentCoordinator.Port == int(in.Port) && serv.currentCoordinator.Ip == in.Ip {
				serv.currentCoordinator = items.LifeGuard{Port: -1, Ip: ""}
				serv.currentLoadBalancer = items.EndPoint{Ip: "", Port: -1}
			}
			return &api_gateway.RemoveLifeGuardNodeResponse{}, nil
		}
	}
	return nil, errors.New("could not remove lifeguard")
}

func (serv *APIGatewayServer) SetLifeGuardCoordinator(
	ctx context.Context, in *api_gateway.SetLifeGuardCoordinatorRequest,
) (*api_gateway.SetLifeGuardCoordinatorResponse, error) {
	if serv.currentCoordinator.Port != -1 || serv.currentCoordinator.Ip != "" {
		return &api_gateway.SetLifeGuardCoordinatorResponse{}, errors.New("coordinator already set")
	}

	lifeGuard, found := (*serv.lifeGuards)[int(in.LifeGuardId)]

	if !found {
		println("Got request to add lifeGuard with invalid ID.")
		return &api_gateway.SetLifeGuardCoordinatorResponse{}, errors.New("there does not exist a lifeGuard with this ID")
	}

	serv.currentCoordinator = lifeGuard
	println("New Coordinator: " + lifeGuard.Ip + ":" + strconv.Itoa(lifeGuard.Port) + " with id: " + strconv.Itoa(int(in.LifeGuardId)))

	serv.currentLoadBalancer = items.EndPoint{Ip: lifeGuard.Ip, Port: int(in.LoadBalancerPort)}
	println("New loadBalancer: " + serv.currentLoadBalancer.Ip + ":" + strconv.Itoa(serv.currentLoadBalancer.Port))

	return &api_gateway.SetLifeGuardCoordinatorResponse{}, nil
}

func (serv *APIGatewayServer) GetLifeGuardCoordinator(
	ctx context.Context, in *api_gateway.GetLifeGuardCoordinatorRequest,
) (*api_gateway.GetLifeGuardCoordinatorResponse, error) {

	return &api_gateway.GetLifeGuardCoordinatorResponse{Ip: serv.currentCoordinator.Ip, Port: int32(serv.currentCoordinator.Port)}, nil
}

func (serv *APIGatewayServer) GetCurrentLoadBalancer(
	ctx context.Context, in *api_gateway.GetCurrentLoadBalancerRequest,
) (*api_gateway.GetCurrentLoadBalancerResponse, error) {
	if serv.currentLoadBalancer.Port == -1 || serv.currentLoadBalancer.Ip == "" {
		return &api_gateway.GetCurrentLoadBalancerResponse{Ip: "", Port: -1}, errors.New("loadBalancer not set")
	}

	return &api_gateway.GetCurrentLoadBalancerResponse{Ip: serv.currentLoadBalancer.Ip, Port: int32(serv.currentLoadBalancer.Port)}, nil
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

	//println("Did not find nextLifeguard for lifeGuard id: " + strconv.Itoa(int(in.LifeGuardId)))
	return &api_gateway.GetNextLifeGuardResponse{Ip: "", Port: int32(-1)}, errors.New("could not find you in list of lifeGuards")
}

func (serv *APIGatewayServer) GetMaxLifeGuards(
	ctx context.Context, in *api_gateway.GetMaxLifeGuardsRequest,
) (*api_gateway.GetMaxLifeGuardsResponse, error) {
	isMax := int(serv.maxLifeGuards) == len(*serv.lifeGuards)
	return &api_gateway.GetMaxLifeGuardsResponse{MaxLifeGuards: serv.maxLifeGuards, IsCurrentlyMaxNumber: isMax}, nil
}
