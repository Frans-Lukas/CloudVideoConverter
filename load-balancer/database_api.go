package video_converter

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"log"
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
	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, constants.ProjectID)
	if err != nil {
		log.Fatalf("datastore client was not accessible")
	}
	return ConversionObjectsClient{Client: *dsClient}
}

func (store *ConversionObjectsClient) AddParts(files []string, count int, conversionType string, token string) {
	ctx := context.Background()
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
		ctx := context.Background()
		_, err := store.Put(ctx, datastore.NameKey(KIND, v, nil), &objectToAdd)
		if err != nil {
			log.Println("failed to add ", v, " to datastore")
		}
	}
}

func (store *ConversionObjectsClient) StartConversionForParts(token string) {
	ctx := context.Background()
	q := datastore.NewQuery(KIND).Filter("Token =", token)
	var objects []ConversionObject
	keys, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", token, " because: ", err.Error())
	}
	for i, v := range objects {
		v.InProgress = true
		v.ConversionStartTime = time.Now()
		_, err := store.Put(ctx, keys[i], &v)
		if err != nil {
			log.Println("failed to add ", keys[i], " to datastore")
		}
	}
}

func (store *ConversionObjectsClient) GetPartsInProgress(token string) []ConversionObject {
	ctx := context.Background()
	q := datastore.NewQuery(KIND).Filter("InProgress =", true)
	var objects []ConversionObject
	_, err := store.GetAll(ctx, q, &objects)
	if err != nil {
		log.Println("could not update conversion status for token ", token, " because: ", err.Error())
	}
	return objects
}