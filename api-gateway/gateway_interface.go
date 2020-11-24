package api_gateway

import (
	"context"
	api_gateway "github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	videoconverter "github.com/Frans-Lukas/cloudvideoconverter/generated"
)

type APIGatewayServer struct {
	api_gateway.UnimplementedAPIGateWayServer
}

func CreateNewServer() APIGatewayServer {
	val := APIGatewayServer{}
	return val
}

func (serv *APIGatewayServer) AddServiceEndpoint(
	ctx context.Context, in *api_gateway.,
) (*videoconverter.UploadTokenResponse, error) {

	rpc
	AddServiceEndpoint(ServiceEndPoint)
	returns (AddedServiceEndPoint){}
	rpc
	DisableServiceEndpoint(DisableServiceEndPoint)
	returns (DisabledServiceEndPoint){}
	rpc
	GetActiveServiceEndpoints(ServiceEndPointsRequest)
	returns (ServiceEndPointList){}
