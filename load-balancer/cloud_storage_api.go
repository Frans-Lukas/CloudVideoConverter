package video_converter

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type StorageClient struct {
	*storage.Client
}

func storageClient() StorageClient {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		println(err.Error())
	}
	storageClient := StorageClient{
		client,
	}
	return storageClient
}

func (cli *StorageClient) getConvertedBucketHandle() *storage.BucketHandle {
	bkt := cli.Bucket(constants.ConvertedVideosBucketName)
	return bkt
}

func (cli *StorageClient) getConvertedVideos() []string {
	bkt := cli.getConvertedBucketHandle()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	objectIterator := bkt.Objects(ctx, nil)

	var names []string
	for {
		attrs, err := objectIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		names = append(names, attrs.Name)
	}
	return names
}

func (cli *StorageClient) DownloadSpecificParts(token string) {
	bkt := cli.getConvertedBucketHandle()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	query := storage.Query{Prefix: token}
	objectIterator := bkt.Objects(ctx, &query)

	for {
		attrs, err := objectIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		rc, err := bkt.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Fatalf("DownloadSpecificParts: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}
		defer rc.Close()
		slurp, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Fatalf("DownloadSpecificParts: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}

		f, err := os.Create(constants.LocalStorage + token + "." + "converted")
		if err != nil {
			log.Fatalf("DownloadSpecificParts, create file: %v", err)
		}
		_, err = f.Write(slurp)
		if err != nil {
			log.Fatalf("DownloadSpecificParts, write to file: %v", err)
		}

		f.Close()

	}
}
