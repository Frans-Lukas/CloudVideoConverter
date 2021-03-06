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
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const tokenLength = 20
const tokenTimeOutSeconds = 60 * 10
const megaByte = 1000000
const sizeLimit = megaByte * 1
const NumberOfMovingAvgsToAccountFor = 15

type VideoConverterServer struct {
	videoconverter.UnimplementedVideoConverterLoadBalancerServer
	ActiveTokens                  *map[string]items.Token
	originalQueueLen              int
	ConversionQueue               *[]ConversionObjectInfo
	ActiveServices                *map[string]VideoConverterClient
	databaseClient                *ConversionObjectsClient
	storageClient                 *StorageClient
	apiGatewayAddress             string
	timeSinceVMCreationOrDeletion *time.Time
	movingAverageList             *[]int
	currentAveragePos             uint32
	name                          string
}

type ConversionObjectInfo struct {
	name       string
	outputType string
}

type VideoConverterClient struct {
	client  videoconverter.VideoConverterServiceClient
	address string
}

func CreateNewServer(name string) VideoConverterServer {
	activeTokens := make(map[string]items.Token, 0)
	conversionQueue := make([]ConversionObjectInfo, 0)
	movingAverage := make([]int, 0)
	activeServices := make(map[string]VideoConverterClient, 0)
	dataBaseClient := NewConversionObjectsClient()
	storageClient := CreateStorageClient()
	timer := time.Time{}
	val := VideoConverterServer{
		ActiveTokens:                  &activeTokens,
		originalQueueLen:              0,
		ConversionQueue:               &conversionQueue,
		ActiveServices:                &activeServices,
		databaseClient:                &dataBaseClient,
		storageClient:                 &storageClient,
		apiGatewayAddress:             "",
		timeSinceVMCreationOrDeletion: &timer,
		movingAverageList:             &movingAverage,
		currentAveragePos:             0,
		name:                          name,
	}
	return val
}

func (serv *VideoConverterServer) UpdateActiveServices(address string) {
	//println("Trying to update active services with address: ", address)

	if address == "" {
		println("ADDRESS NOT SET!")
		return
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(time.Second*4))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	//println("connected")
	apiGateway := api_gateway.NewAPIGateWayClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//println("requesting service endpoints:")
	endPoints, err := apiGateway.GetActiveServiceEndpoints(ctx, &api_gateway.ServiceEndPointsRequest{})
	*serv.ActiveServices = make(map[string]VideoConverterClient)
	for _, v := range endPoints.EndPoint {
		address := v.Ip + ":" + strconv.Itoa(int(v.Port))
		//println("got service endpoint: ", address)
		if _, ok := (*serv.ActiveServices)[address]; !ok {
			(*serv.ActiveServices)[address] = makeServiceConnection(address)
		}
	}
	//println("done updating service endpoints")
}

func (serv *VideoConverterServer) PollActiveServices(address string) {
	unresponsiveClients := make([]string, 0)
	for addr, v := range *serv.ActiveServices {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := v.client.IsAlive(ctx, &videoconverter.IsAliveRequest{})
		if err != nil {
			log.Println("unresponsive client: " + v.address + " " + err.Error())
			unresponsiveClients = append(unresponsiveClients, addr)
		}
	}

	for _, v := range unresponsiveClients {
		println("deleting unresponsive client: ", v)
		delete(*serv.ActiveServices, v)
		notifyAPIGatewayOfDeadClient(v, serv.apiGatewayAddress)
	}
}

func notifyAPIGatewayOfDeadClient(removeAddr string, apiAddress string) {
	println("trying to connect to APIGateway: ", apiAddress)
	conn, err := grpc.Dial(apiAddress, grpc.WithInsecure(), grpc.WithTimeout(time.Second*3))
	if err != nil {
		log.Println("did not connect to api gateway: %v", err)
		return
	}
	defer conn.Close()
	println("connected")
	apiGateway := api_gateway.NewAPIGateWayClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ip := strings.Split(removeAddr, ":")[0]
	portString := strings.Split(removeAddr, ":")[1]
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Println("Failed to split address: " + removeAddr)
		return
	}
	_, err = apiGateway.DisableServiceEndpoint(ctx, &api_gateway.DisableServiceEndPointRequest{Ip: ip, Port: int32(port)})
	if err != nil {
		println("failed to disable service endpoint", err.Error())
	}
}

func makeServiceConnection(address string) VideoConverterClient {
	//println("connecting to service endpoint: ", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(time.Second*3))
	//println("connected to service endpoint!")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//println("connected")
	return VideoConverterClient{client: videoconverter.NewVideoConverterServiceClient(conn), address: address}
}

func (serv *VideoConverterServer) RequestUploadToken(ctx context.Context, in *videoconverter.UploadTokenRequest) (*videoconverter.UploadTokenResponse, error) {
	tokenString := GenerateRandomString()
	println("received request for upload token, responding with: ", tokenString)
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

func saveFile(fileName string, imageBytes *bytes.Buffer) error {
	//dir, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(dir)

	imagePath := constants.LocalStorage + fileName + ".mp4"
	println("saveFile imagePath: ", imagePath)
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

func (serv *VideoConverterServer) WorkManagementLoop() {
	var count uint = 0
	for {
		// Update Clients
		//println("updating services and queue")
		serv.UpdateActiveServices(serv.apiGatewayAddress)
		serv.PollActiveServices(serv.apiGatewayAddress)
		serv.handleQueueFromDB()
		if count%20 == 0 {
			PrintKeyValue("numberOfActiveServices", len(*serv.ActiveServices))
			PrintKeyValue("sizeOfQueue", len(*serv.ConversionQueue))
			PrintCPUUsage(serv.name)
		}

		// Handle Videos
		//println("sending work to clients")
		serv.SendWorkToClients()
		//println("checking if can merge")
		tokens := serv.databaseClient.CheckForMergeableFiles()
		if len(tokens) > 0 {
			for _, token := range tokens {
				if convertedFileExists(token) {
					println(token, " is already merged, skipping.")
				} else {
					serv.downloadAndMergeFiles(token)
					println("conversion for ", token, " is done and merged!")
				}
				if _, ok := (*serv.ActiveTokens)[token]; ok {
					*(*serv.ActiveTokens)[token].ConversionDone = true
				}
			}
		}

		// Handle Clients
		count++
		time.Sleep(constants.WorkManagementLoopSleepTime)
	}
}

func (serv *VideoConverterServer) SendWorkToClients() {
	for addr, client := range *serv.ActiveServices {
		if len(*serv.ConversionQueue) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		//println("Checking if ", addr, " can work")
		response, err := client.client.AvailableForWork(ctx, &videoconverter.AvailableForWorkRequest{})
		if err != nil {
			println(" conv check err: ", err.Error())
			return
		}
		if response.AvailableForWork {
			println("sending work to ", addr)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			nextJob := (*serv.ConversionQueue)[0]
			*serv.ConversionQueue = (*serv.ConversionQueue)[1:]
			serv.databaseClient.SetConversionAddressForPart(nextJob.name, client.address)
			client.client.StartConversion(ctx, &videoconverter.ConversionRequest{Token: nextJob.name, OutputType: nextJob.outputType})
		}
	}
}

func (serv *VideoConverterServer) StartConversion(ctx context.Context, in *videoconverter.ConversionRequest) (*videoconverter.ConversionResponse, error) {
	err, filesToConvert := serv.databaseClient.SetConversionTypeForParts(in.Token, in.OutputType)
	println("load balancer starting conversion for ", in.Token)
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

func (serv *VideoConverterServer) conversionIsNotFinishedForToken(token string) bool {
	// conversion is not finished
	return !*(*serv.ActiveTokens)[token].ConversionDone
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

	err = saveFile(tokenString, &imageData)
	if err != nil {
		return err
	}

	err = splitVideo(tokenString)
	if err != nil {
		return err
	}
	//deleteFullVideo(tokenString)
	sendVideosToCloudStorage(tokenString)
	serv.sendVideoInformationToDatabase(tokenString)

	//mergeVideo(tokenString)

	// ...

	return nil
}

func (serv *VideoConverterServer) sendVideoInformationToDatabase(token string) {
	fileNames, err := getVideoParts(token)
	if err != nil {
		log.Println("failed to get video parts, " + err.Error())
	}
	serv.databaseClient.AddParts(fileNames, len(fileNames), "mkv", token)
}

func sendVideosToCloudStorage(token string) {
	println("uploading files to cloud storage, token: ", token)
	fileNames, err := getVideoParts(token)
	if err != nil {
		log.Println("failed to get video parts, " + err.Error())
	}
	uploadFiles(fileNames)
	println("files uploaded to cloud storage")

}

func uploadFiles(fileNames []string) {
	storageClient := CreateStorageClient()
	storageClient.listBuckets()
	println(constants.UnconvertedVideosBucketName)
	for _, fileName := range fileNames {
		storageClient.UploadUnconvertedPart(fileName)
	}
}

func (server *VideoConverterServer) tokenIsInvalid(token string) bool {
	if tokenCreationTime, ok := (*server.ActiveTokens)[token]; ok {
		println("Time since token was created: ", time.Since(*tokenCreationTime.CreationTime).Seconds())
		println("in seconds: ", tokenTimeOutSeconds)
		if time.Since(*tokenCreationTime.CreationTime).Seconds() < tokenTimeOutSeconds {
			return false
		}
	}
	return true
}

func (serv *VideoConverterServer) Download(request *videoconverter.DownloadRequest, stream videoconverter.VideoConverterLoadBalancer_DownloadServer) error {

	token := request.Id
	if serv.tokenIsInvalid(token) {
		return errors.New("token is invalid or has timed out: " + request.Id)
	}

	//TODO load corresponding file from directory
	file, err := os.Open(constants.LocalStorage + token + constants.FinishedConversionExtension)

	if err != nil {
		log.Fatalf("Download, Open failed: %v", err)
	}
	defer file.Close()

	buf := make([]byte, constants.DownloadChunkSizeInBytes)

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

func (serv *VideoConverterServer) MarkTokenAsComplete(ctx context.Context, in *videoconverter.MarkTokenAsCompleteRequest) (*videoconverter.MarkTokenAsCompleteResponse, error) {
	DeleteFiles(in.Token)
	serv.storageClient.DeleteConvertedParts(in.Token)
	serv.databaseClient.DeleteWithToken(in.Token)
	return &videoconverter.MarkTokenAsCompleteResponse{}, nil
}

func DeleteFiles(prefix string) {
	files, err := filepath.Glob("/home/group9/CloudVideoConverter/localStorage/" + prefix + "*")
	if err != nil {
		panic(err)
	}
	println("trying to delete ", len(files), " number of files.")
	for _, f := range files {
		println("removing file: ", f)
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
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
	//TODO: check this is not null

	if _, ok := (*serv.ActiveTokens)[in.StatusId]; ok {
		if *(*serv.ActiveTokens)[in.StatusId].ConversionDone {
			return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Done}, nil
		}
		if *(*serv.ActiveTokens)[in.StatusId].ConversionStarted {
			if *(*serv.ActiveTokens)[in.StatusId].ConversionFailed {
				return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Failed}, nil
			}
			return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_InProgress}, nil
		}
	} else {
		return &videoconverter.ConversionStatusResponse{Code: videoconverter.ConversionStatusCode_Failed}, nil
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
		println("Checking active tokens: ")
		keysToDelete := make([]string, 0)
		for token, _ := range *serv.ActiveTokens {
			println("checking if I can delete token: ", token)
			if serv.tokenIsInvalid(token) {
				println("deleting invalid token from everywhere.")
				serv.deleteTokenFromEverywhere(token)
				keysToDelete = append(keysToDelete, token)
				break
			}
		}
		for _, key := range keysToDelete {
			delete(*serv.ActiveTokens, key)
		}

		_, parts := serv.databaseClient.GetFinishedParts()
		for _, v := range parts {
			println("time since: ", time.Since(v.ConversionStartTime).Seconds())
			println("tokenTimeout: ", tokenTimeOutSeconds)
			if time.Since(v.ConversionStartTime).Seconds() > tokenTimeOutSeconds {
				serv.deleteTokenFromEverywhere(v.Token)
			}
		}

		//fmt.Printf("map after deletion: %v", *serv.ActiveConversions)

		time.Sleep(time.Second * 5)
	}
}

func (serv *VideoConverterServer) deleteTokenFromEverywhere(token string) {
	DeleteFiles(token)
	serv.databaseClient.DeleteWithToken(token)
	serv.storageClient.DeleteUnconvertedParts(token)
	serv.storageClient.DeleteConvertedParts(token)
}

func (serv *VideoConverterServer) downloadAndMergeFiles(token string) {
	serv.storageClient.DownloadConvertedParts(token)
	err := mergeVideo(token)
	if err != nil {
		println(err.Error())
	}
}

func (serv *VideoConverterServer) SetApiGatewayAddress(address string) {
	serv.apiGatewayAddress = address
}

func (serv *VideoConverterServer) ManageClients() {
	for {
		if serv.shouldReduceNumberOfServices() {
			serv.reduceNumberOfServices()
		} else if serv.shouldIncreaseNumberOfServices() {
			diff := len(*serv.ConversionQueue) - len(*serv.ActiveServices)
			serv.IncreaseNumberOfServices(int(math.Min(float64(diff), 3)))
		}
		time.Sleep(time.Second * 4)
	}
}

func (serv *VideoConverterServer) shouldReduceNumberOfServices() bool {
	serv.addToMovingAverageList(len(*serv.ActiveServices) - len(*serv.ConversionQueue))
	// for real life scenarios, this check should be changed as we could receive ONE gigantic job that creates many vms.
	if serv.currentAveragePos < 5 {
		return false
	}
	sum := 0
	for _, v := range *serv.movingAverageList {
		sum += v
	}
	sum = sum / len(*serv.movingAverageList)
	println("moving average: ", sum)
	return sum >= 2 && time.Since(*serv.timeSinceVMCreationOrDeletion) > time.Second*30
}

func (serv *VideoConverterServer) shouldIncreaseNumberOfServices() bool {
	diff := len(*serv.ConversionQueue) - len(*serv.ActiveServices)
	return diff > 0
}

func (serv *VideoConverterServer) countNonFinishedConversions() int {
	count := 0
	for token, _ := range *serv.ActiveTokens {
		if serv.conversionIsNotFinishedForToken(token) {
			count += 1
		}
	}
	return count
}

func (serv *VideoConverterServer) reduceNumberOfServices() {
	println("Attempting shutdown of service")
	for i, v := range *serv.ActiveServices {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := v.client.AvailableForWork(ctx, &videoconverter.AvailableForWorkRequest{})
		if err != nil {
			println("service did not respond")
			continue
		}
		if response.AvailableForWork {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := v.client.ShutDown(ctx, &videoconverter.ShutDownRequest{})
			if err != nil {
				println("failed to shutdown client ", i)
			} else {
				println("Shutdown service")
				serv.resetVMTimer()
				return
			}
		}
	}
	println("No service was shut down")
}

func (serv *VideoConverterServer) IncreaseNumberOfServices(numToCreate int) {
	println("Starting new service")
	scriptPath := "/home/group9/CloudVideoConverter/scripts/tfScripts/Service/startServiceVM.sh"
	numberOfVms := strconv.Itoa(len(*serv.ActiveServices) + numToCreate)
	cmd := exec.Command(scriptPath, numberOfVms)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Println("could not increaseNumberOfServices: " + out.String())
		log.Println(err.Error(), ": ", stderr.String())
	} else {
		log.Println(out.String())
	}
}

func (serv *VideoConverterServer) enoughTimeSinceVMCreationOrDeletion() bool {
	//println("Time till VM can be created or deleted: " + fmt.Sprintf("%f", 60-time.Since(*serv.timeSinceVMCreationOrDeletion).Seconds()))
	return time.Since(*serv.timeSinceVMCreationOrDeletion).Seconds() > constants.SecondsBetweenVMCreationAndDeletion
}

func (serv *VideoConverterServer) resetVMTimer() {
	now := time.Now()
	serv.timeSinceVMCreationOrDeletion = &now
}

func (serv *VideoConverterServer) handleQueueFromDB() {
	keys, unfinishedParts := serv.databaseClient.GetUnfinishedParts()
	newConversionQueue := make([]ConversionObjectInfo, 0)
	serv.originalQueueLen = len(unfinishedParts)
	for index, v := range unfinishedParts {
		if _, ok := (*serv.ActiveServices)[v.ConverterAddress]; !ok && v.InProgress && !v.Done {
			serv.databaseClient.MarkConversionAsNotInProgress(keys[index].Name)
		} else if v.InProgress == false && !v.Done {
			newConversionQueue = append(newConversionQueue, ConversionObjectInfo{
				name:       keys[index].Name,
				outputType: v.ConversionType,
			})
			creationTime := v.ConversionStartTime
			isStarted := false
			isDone := false
			isFailed := false
			(*serv.ActiveTokens)[v.Token] = items.Token{
				CreationTime:      &creationTime,
				ConversionStarted: &isStarted,
				ConversionDone:    &isDone,
				ConversionFailed:  &isFailed,
			}
		}
	}
	serv.ConversionQueue = &newConversionQueue
}
func (serv *VideoConverterServer) addToMovingAverageList(rest int) {
	if serv.currentAveragePos >= NumberOfMovingAvgsToAccountFor {
		(*serv.movingAverageList)[serv.currentAveragePos%NumberOfMovingAvgsToAccountFor] = rest
	} else {
		*serv.movingAverageList = append(*serv.movingAverageList, rest)
	}
	serv.currentAveragePos++
	if serv.currentAveragePos == math.MaxUint32 {
		serv.currentAveragePos = NumberOfMovingAvgsToAccountFor
	}
}
