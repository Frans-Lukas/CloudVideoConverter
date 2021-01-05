package video_converter

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"github.com/Frans-Lukas/cloudvideoconverter/helpers"
	"log"
	"strings"
	"time"
)

type ConversionObject struct {
	ConversionStartTime time.Time
	ConversionType      string
	ConverterAddress    string
	Done                bool
	InProgress          bool
	PartCount           int
	Token               string
}

const KIND = "Job"

type ConversionObjectsClient struct {
	datastore.Client
}

func NewConversionObjectsClient() ConversionObjectsClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	dsClient, err := datastore.NewClient(ctx, constants.ProjectID)
	if err != nil {
		log.Fatalf("datastore client was not accessible, ", err.Error())
	}
	return ConversionObjectsClient{Client: *dsClient}
}

func (store *ConversionObjectsClient) AddParts(files []string, count int, conversionType string, token string) {
	for _, v := range files {
		objectToAdd := ConversionObject{
			ConversionStartTime: time.Now(),
			ConversionType:      conversionType,
			ConverterAddress:    "",
			Done:                false,
			InProgress:          false,
			PartCount:           len(files),
			Token:               token,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err := store.Put(ctx, datastore.NameKey(KIND, v, nil), &objectToAdd)
		if err != nil {
			log.Println("failed to add ", v, " to datastore")
		}
	}
}
func (store *ConversionObjectsClient) DeleteWithToken(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND).Filter("Token =", token)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", token, " because: ", err.Error())
		return
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	store.DeleteMulti(ctx, keys)
}

func (store *ConversionObjectsClient) MarkConversionAsDone(file string) (error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	correctFile := helpers.ChangeFileExtension(file, "mp4")
	println("correctFile: ", correctFile)
	q := datastore.NewQuery(KIND).Filter("__key__ =", datastore.NameKey(KIND, correctFile, nil))
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", correctFile, " because: ", err.Error())
		return err
	}
	for i, object := range objects {
		object.Done = true
		_, err := store.Put(ctx, keys[i], &object)
		if err != nil {
			log.Println("failed to add ", keys[i], " to datastore")
			return err
		}
	}
	return nil
}

func (store *ConversionObjectsClient) StartConversionForParts(token string, outputType string) (error, *[]ConversionObjectInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND).Filter("Token =", token)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", token, " because: ", err.Error())
		return err, nil
	}
	fileNames := make([]ConversionObjectInfo, 0)
	for i, object := range objects {
		object.InProgress = true
		object.ConversionType = outputType
		object.ConversionStartTime = time.Now()
		fileNames = append(fileNames, ConversionObjectInfo{keys[i].Name, object.ConversionType})
		_, err := store.Put(ctx, keys[i], &object)
		if err != nil {
			log.Println("failed to add ", keys[i], " to datastore")
			return err, nil
		}
	}

	return nil, &fileNames
}
func (store *ConversionObjectsClient) RestartConversionForParts(token string) (error, *[]ConversionObjectInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND).Filter("Token =", token)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", token, " because: ", err.Error())
		return err, nil
	}
	fileNames := make([]ConversionObjectInfo, 0)
	for i, object := range objects {
		object.InProgress = false
		fileNames = append(fileNames, ConversionObjectInfo{keys[i].Name, object.ConversionType})
		_, err := store.Put(ctx, keys[i], &object)
		if err != nil {
			log.Println("failed to add ", keys[i], " to datastore")
			return err, nil
		}
	}

	return nil, &fileNames
}

func (store *ConversionObjectsClient) GetPartsInProgress() []ConversionObjectInfo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND).Filter("InProgress =", true)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("Could not get parts in progress")
	}
	fileNames := make([]ConversionObjectInfo, 0)
	for i, object := range objects {
		fileNames = append(fileNames, ConversionObjectInfo{keys[i].Name, object.ConversionType})
	}
	return fileNames
}

func (store *ConversionObjectsClient) CheckForMergeableFiles() []string {
	keys, objects := store.GetFinishedParts()
	//println("found ", len(keys), " finished parts")

	countMap := make(map[string][]ConversionObjectInfo, 0)
	desiredCountMap := make(map[string]int, 0)

	for i, v := range keys {
		token := strings.Split(v.Name, "-")[0]
		//println(v.Name)
		//println("token used for counting: ", token)
		if _, ok := countMap[token]; !ok {
			countMap[token] = make([]ConversionObjectInfo, 0)
		}

		countMap[token] = append(countMap[token], ConversionObjectInfo{v.Name, objects[i].ConversionType})
		desiredCountMap[token] = objects[i].PartCount
	}

	output := make([]string, 0)
	for i, v := range desiredCountMap {
		//println("seeing if part is done v: ", v, " part count: ", len(countMap[i]))
		if v == len(countMap[i]) {
			output = append(output, i)
			//println("appedning finished part: ", i)
		}
	}
	return output
}

func (store *ConversionObjectsClient) GetFinishedParts() ([]*datastore.Key, []ConversionObject) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND).Filter("Done =", true)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("Could not get finished parts")
	}
	return keys, objects
}

func (store *ConversionObjectsClient) DeleteAllEntities() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	q := datastore.NewQuery(KIND)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("Could not get all converted datastore enteties", err.Error())
		return
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	store.DeleteMulti(ctx, keys)
}
