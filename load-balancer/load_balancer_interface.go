package video_converter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Frans-Lukas/cloudvideoconverter/api-gateway/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/generated"
	"github.com/Frans-Lukas/cloudvideoconverter/load-balancer/server/items"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const tokenLength = 20
const tokenTimeOutSeconds = 60 * 20
const megaByte = 1000000
const sizeLimit = megaByte * 1

type VideoConverterServer struct {
	videoconverter.UnimplementedVideoConverterLoadBalancerServer
	ActiveTokens    *map[string]items.Token
	ConversionQueue *[]string
	ActiveServices  *map[string]VideoConverterClient
	databaseClient  *ConversionObjectsClient
}

type VideoConverterClient struct {
	client  videoconverter.VideoConverterServiceClient
	address string
}

func CreateNewServer() VideoConverterServer {
	activeTokens := make(map[string]items.Token, 0)
	conversionQueue := make([]string, 0)
	activeServices := make(map[string]VideoConverterClient, 0)
	dataBaseClient := NewConversionObjectsClient()
	val := VideoConverterServer{
		ActiveTokens:    &activeTokens,
		ConversionQueue: &conversionQueue,
		ActiveServices:  &activeServices,
		databaseClient:  &dataBaseClient,
	}
	return val
}

func (serv *VideoConverterServer) UpdateActiveServices(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	apiGateway := api_gateway.NewAPIGateWayClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	endPoints, err := apiGateway.GetActiveServiceEndpoints(ctx, &api_gateway.ServiceEndPointsRequest{})
	*serv.ActiveServices = make(map[string]VideoConverterClient)
	for _, v := range endPoints.EndPoint {
		address := strconv.Itoa(int(v.Port)) + ":" + v.Ip
		if _, ok := (*serv.ActiveServices)[address]; !ok {
			(*serv.ActiveServices)[address] = makeServiceConnection(address)
		}
	}
}

func (serv *VideoConverterServer) PollActiveServices(address string) {
	unresponsiveClients := make([]string, 0)
	for i, v := range *serv.ActiveServices {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := v.client.IsAlive(ctx, &videoconverter.IsAliveRequest{})
		if err != nil {
			log.Println("unresponsive client: " + v.address + " " + err.Error())
			unresponsiveClients = append(unresponsiveClients, i)
		}
	}

	for _, v := range unresponsiveClients {
		delete(*serv.ActiveServices, v)
		notifyAPIGatewayOfDeadClient(v)
	}
}
func notifyAPIGatewayOfDeadClient(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	apiGateway := api_gateway.NewAPIGateWayClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ip := strings.Split(address, ":")[0]
	portString := strings.Split(address, ":")[1]
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Println("Failed to split address: " + address)
		return
	}
	apiGateway.DisableServiceEndpoint(ctx, &api_gateway.DisableServiceEndPoint{Ip: ip, Port: int32(port)})
}

func makeServiceConnection(address string) VideoConverterClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	println("connected")
	return VideoConverterClient{client: videoconverter.NewVideoConverterServiceClient(conn), address: address}
}

func (serv *VideoConverterServer) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	creationTime := time.Now()
	isStarted := false
	isDone := false
	isFailed := false
	(*serv.ActiveTokens)[tokenString] = items.Token{
		CreationTime:      &creationTime,
		ConversionStarted: &isStarted,
		ConversionDone:    &isDone,
		ConversionFailed:  &isFailed,
	}
	return &videoconverter.UploadTokenResponse{Token: tokenString}, nil
}

func saveImage(fileName string, imageBytes *bytes.Buffer) error {
	imagePath := constants.LocalStorage + fileName + ".mp4"
	file, err := os.Create(imagePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}
	_, err = imageBytes.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write image to file: %w", err)
	}
	return nil
}

func (serv *VideoConverterServer) SendWorkLoop() {
	for {
		time.Sleep(time.Millisecond * 100)
		serv.SendWorkToClients()
	}
}

func (serv *VideoConverterServer) SendWorkToClients() {
	for _, client := range *serv.ActiveServices {
		if len(*serv.ConversionQueue) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := client.client.AvailableForWork(ctx, &videoconverter.AvailableForWorkRequest{})
		if err != nil {
			println(" conv check err: ", err.Error())
		}
		if response.AvailableForWork {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			nextJob := (*serv.ConversionQueue)[0]
			*serv.ConversionQueue = (*serv.ConversionQueue)[1:]
			client.client.StartConversion(ctx, &videoconverter.ConversionRequest{Token: nextJob})
		}
	}
}

func (serv *VideoConverterServer) LoadQueueFromDB() {
	filesToConvert := serv.databaseClient.GetPartsInProgress()
	for _, v := range filesToConvert {
		*serv.ConversionQueue = append(*serv.ConversionQueue, v)
	}
}

func (serv *VideoConverterServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	err, filesToConvert := serv.databaseClient.StartConversionForParts(in.Token, in.OutputType)
	if err != nil {
		return &videoconverter.ConversionResponse{}, err
	}
	for _, v := range *filesToConvert {
		*serv.ConversionQueue = append(*serv.ConversionQueue, v)
	}
	return &videoconverter.ConversionResponse{}, nil
}

func (serv *VideoConverterServer) conversionIsInProgressForToken(token string) bool {
	// conversion is started but not done or failed
	return *(*serv.ActiveTokens)[token].ConversionStarted && !*(*serv.ActiveTokens)[token].ConversionDone && !*(*serv.ActiveTokens)[token].ConversionFailed
}

func (serv *VideoConverterServer) performConversion(app string, arg0 string, arg1 string, arg2 string, in *videoconverter.ConversionRequest) {
	cmd := exec.Command(app, arg0, arg1, arg2)
	*(*serv.ActiveTokens)[in.Token].ConversionStarted = true
	err := cmd.Run()
	if err != nil {
		*(*serv.ActiveTokens)[in.Token].ConversionFailed = true
		return
	} else {
		*(*serv.ActiveTokens)[in.Token].ConversionDone = true
	}
	file, err := os.Open(arg2)
	defer file.Close()
	if err != nil {
		println(err.Error())
		return
	}
	os.Rename(arg2, constants.LocalStorage+in.Token)
}

func (serv *VideoConverterServer) Upload(stream videoconverter.VideoConverterLoadBalancer_UploadServer) error {

	imageData := bytes.Buffer{}
	tokenString := ""

	for {
		streamData, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			}

			err = errors.New("failed unexpectadely while reading chunks from stream")
			return err
		}
		switch streamData.RequestType.(type) {
		case *videoconverter.Chunk_Content:
			if tokenString == "" {
				return errors.New("token must be first message")
			}
			chunk := streamData.GetContent()
			imageData.Write(chunk)
		case *videoconverter.Chunk_Token:
			token := streamData.GetToken()
			if serv.tokenIsInvalid(token) {
				return errors.New("token is invalid or has timed out: " + token)
			}
			tokenString = token
		}
	}

	// once the transmission finished, send the
	// confirmation if nothign went wrong
	err := stream.SendAndClose(&videoconverter.UploadStatus{
		RetrievalToken: tokenString,
	})

	err = saveImage(tokenString, &imageData)
	if err != nil {
		return err
	}

	splitVideo(tokenString)
	mergeVideo(tokenString)

	// ...

	return nil
}

func (server *VideoConverterServer) tokenIsInvalid(token string) bool {
	if tokenCreationTime, ok := (*server.ActiveTokens)[token]; ok {
		if time.Since(*tokenCreationTime.CreationTime).Seconds() < tokenTimeOutSeconds {
			return false
		}
	}
	return true
}

func (serv *VideoConverterServer) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverterLoadBalancer_DownloadServer) error {
	if serv.tokenIsInvalid(request.Id) {
		return errors.New("token is invalid or has timed out: " + request.Id)
	}
	//TODO set chunksize in a global way
	chunksize := 1000

	//TODO check if id is valid
	id := request.Id

	//TODO load corresponding file from directory
	file, err := os.Open(constants.LocalStorage + id)
	if err != nil {
		log.Fatalf("Download, Open failed: %v", err)
	}

	buf := make([]byte, chunksize)

	for {
		n, err := file.Read(buf)

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Download, Read: %v", err)
		}

		stream.Send(&videoconverter.Chunk{
			RequestType: &videoconverter.Chunk_Content{Content: buf[:n]},
		})
	}

	return nil
}

func (serv *VideoConverterServer) Delete(ctx context.Context, in *videoconverter.DeleteRequest) (*videoconverter.DeleteResponse, error) {
	if serv.tokenIsInvalid(in.Id) {
		return nil, errors.New("token is invalid or has timed out: " + in.Id)
	}
	filePath := constants.LocalStorage + in.Id + ".mp4"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("video to delete does not exist")
	}
	err := os.Remove(filePath)
	if err != nil {
		print(err.Error())
	}
	filePath = constants.LocalStorage + in.Id
	_, err = os.Stat(filePath)
	if err == nil {
		err = os.Remove(filePath)
		if err != nil {
			print(err.Error())
		}
	}
	return &videoconverter.DeleteResponse{}, nil
}

func (serv *VideoConverterServer) ConversionStatus(ctx context.Context, in *videoconverter.ConversionStatusRequest) (*videoconverter.ConversionStatusResponse, error) {
	if *(*serv.ActiveTokens)[in.StatusId].ConversionDone {
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Done}, nil
	}
	if *(*serv.ActiveTokens)[in.StatusId].ConversionStarted {
		if *(*serv.ActiveTokens)[in.StatusId].ConversionFailed {
			return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Failed}, nil
		}
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_InProgress}, nil
	}
	return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_NotStarted}, nil
}
func (serv *VideoConverterServer) resetConversionStatus(token string) {
	*(*serv.ActiveTokens)[token].ConversionStarted = false
	*(*serv.ActiveTokens)[token].ConversionFailed = false
	*(*serv.ActiveTokens)[token].ConversionDone = false
}

func GenerateRandomString() string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < tokenLength; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String() // E.g. "ExcbsVQs"
}

func (serv *VideoConverterServer) DeleteTimedOutVideosLoop() {
	for {
		for token, _ := range *serv.ActiveTokens {
			if serv.tokenIsInvalid(token) {
				filePath := constants.LocalStorage + token + ".mp4"
				_, err := os.Stat(filePath)
				if err == nil {
					println("deleting " + filePath)
					err := os.Remove(filePath)
					if err != nil {
						println(err.Error())
					}
				}
				filePath = constants.LocalStorage + token
				_, err = os.Stat(filePath)
				if err == nil {
					println("deleting " + filePath)
					err := os.Remove(filePath)
					if err != nil {
						println(err.Error())
					}
				}
			}
		}
		time.Sleep(time.Second * 5)
	}
}
