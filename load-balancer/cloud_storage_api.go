package video_converter

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/iterator"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type StorageClient struct {
	*storage.Client
}

func CreateStorageClient() StorageClient {
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

func (cli *StorageClient) getUnconvertedBuketHandle() *storage.BucketHandle {
	bkt := cli.Bucket(constants.UnconvertedVideosBucketName)
	return bkt
}

func (cli *StorageClient) getConvertedVideos() []string {
	bkt := cli.getConvertedBucketHandle()
	return cli.getVideos(bkt)
}

func (cli *StorageClient) getUnconvertedVideos() []string {
	bkt := cli.getUnconvertedBuketHandle()
	return cli.getVideos(bkt)
}

func (cli *StorageClient) getVideos(bkt *storage.BucketHandle) []string {
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

func (cli *StorageClient) UploadConvertedPart(fileName string) {
	bkt := cli.getConvertedBucketHandle()
	cli.uploadFile(bkt, fileName)
}

func (cli *StorageClient) UploadUnconvertedPart(fileName string) {
	bkt := cli.getUnconvertedBuketHandle()
	cli.uploadFile(bkt, fileName)
}

func (cli *StorageClient) uploadFile(bkt *storage.BucketHandle, fileName string) {
	//open local file
	f, err := os.Open(constants.LocalStorage + fileName)
	i, _ := f.Stat()
	println("size of file: " + strconv.Itoa(int(i.Size())))

	if err != nil {
		log.Fatalf("failed to open local file before uploading: " + err.Error())
	}

	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	obj := bkt.Object(fileName)

	w := obj.NewWriter(ctx)
	if _, err = io.Copy(w, f); err != nil {
		log.Fatalf("io.Copy: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Writer.Close: %v", err)
	}
	println("file uploaded!")
}

func (cli *StorageClient) DownloadUnconvertedPart(token string) {
	bkt := cli.getUnconvertedBuketHandle()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	rc, err := bkt.Object(token).NewReader(ctx)
	if err != nil {
		log.Fatalf("DownloadUnconvertedPart: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, token, err)
		return
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatalf("DownloadUnconvertedPart: ioutil.ReadAll: %v", err)
		return
	}

	f, err := os.Create(constants.LocalStorage + token)
	if err != nil {
		log.Fatalf("DownloadUnconvertedPart, create file: %v", err)
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatalf("DownloadUnconvertedPart, write to file: %v", err)
	}

	f.Close()
}

func (cli *StorageClient) DownloadSpecificParts(token string) {
	println("using token: " + token)
	bkt := cli.getConvertedBucketHandle()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	query := storage.Query{Prefix: token}
	objectIterator := bkt.Objects(ctx, &query)

	for {
		attrs, err := objectIterator.Next()
		if err == iterator.Done {
			println("iterator done")
			break
		}
		println("downloading part")
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		rc, err := bkt.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Fatalf("DownloadSpecificParts: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}
		defer rc.Close()
		slurp, err := ioutil.ReadAll(rc)
		println("read " + strconv.Itoa(len(slurp)) + " bytes from downloaded file")
		if err != nil {
			log.Fatalf("DownloadSpecificParts: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}

		f, err := os.Create(constants.LocalStorage + attrs.Name + ".converted")
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

// implicit uses Application Default Credentials to authenticate.
func ImplicitAuth(projectID string) {
	ctx := context.Background()

	// For API packages whose import path is starting with "cloud.google.com/go",
	// such as cloud.google.com/go/storage in this case, if there are no credentials
	// provided, the client library will look for credentials in the environment.
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		println("new client failed")
		log.Fatal(err)
	}

	println("using bucketid: " + projectID)

	it := storageClient.Buckets(ctx, projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			println("buckets failed")
			log.Fatal(err)
		}
		fmt.Println(bucketAttrs.Name)
	}

	// For packages whose import path is starting with "google.golang.org/api",
	// such as google.golang.org/api/cloudkms/v1, use NewService to create the client.
	kmsService, err := cloudkms.NewService(ctx)
	if err != nil {
		println("kms service failed!")
		log.Fatal(err)
	}

	_ = kmsService
}
