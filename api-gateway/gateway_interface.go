package api_gateway

import (
	"context"
	api_gateway "github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
)

type APIGatewayServer struct {
	api_gateway.UnimplementedAPIGateWayServer
}

func CreateNewServer() APIGatewayServer {
	val := APIGatewayServer{}
	return val
}

func (serv *APIGatewayServer) AddServiceEndpoint(
	ctx context.Context, in *api_gateway.ServiceEndPoint,
) (*api_gateway.AddedServiceEndPoint, error) {
	return nil, nil
}

func (serv *APIGatewayServer) DisableServiceEndPoint(
	ctx context.Context, in *api_gateway.DisableServiceEndPoint,
) (*api_gateway.DisabledServiceEndPoint, error) {
	return nil, nil
}
func (serv *APIGatewayServer) GetActiveServiceEndpoints(
	ctx context.Context, in *api_gateway.ServiceEndPointsRequest,
) (*api_gateway.ServiceEndPointList, error) {
	return nil, nil
}