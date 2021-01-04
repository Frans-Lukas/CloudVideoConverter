package api_gateway

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/items"
	"strconv"
)

type APIGatewayServer struct {
	api_gateway.UnimplementedAPIGateWayServer
	endPoints *map[items.EndPoint]bool
}

func CreateNewServer() APIGatewayServer {
	endpoints := make(map[items.EndPoint]bool, 0)
	val := APIGatewayServer{endPoints: &endpoints}
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
	println("removing service from gateway, addr: ", in.Ip, ":", in.Port)
	newEndPoint := items.EndPoint{
		Ip:   in.Ip,
		Port: int(in.Port),
	}
	if _, ok := (*serv.endPoints)[newEndPoint]; ok {
		(*serv.endPoints)[newEndPoint] = false
	} else {
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
