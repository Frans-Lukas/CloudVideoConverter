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
	"strings"
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

func (cli *StorageClient) getSampleVideosBucketHandle() *storage.BucketHandle {
	bkt := cli.Bucket(constants.SampleVideosBucketName)
	return bkt
}

func (cli *StorageClient) getUnconvertedBucketHandle() *storage.BucketHandle {
	bkt := cli.Bucket(constants.UnconvertedVideosBucketName)
	return bkt
}

func (cli *StorageClient) getSampleVideos() []string {
	bkt := cli.getSampleVideosBucketHandle()
	return cli.getVideos(bkt)
}

func (cli *StorageClient) getConvertedVideos() []string {
	bkt := cli.getConvertedBucketHandle()
	return cli.getVideos(bkt)
}

func (cli *StorageClient) getUnconvertedVideos() []string {
	bkt := cli.getUnconvertedBucketHandle()
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
	bkt := cli.getUnconvertedBucketHandle()
	cli.uploadFile(bkt, fileName)
}

func (cli *StorageClient) DeleteConvertedPart(fileName string) {
	bkt := cli.getConvertedBucketHandle()
	cli.deleteFile(bkt, fileName)
}

func (cli *StorageClient) DeleteUnconvertedPart(fileName string) {
	bkt := cli.getUnconvertedBucketHandle()
	cli.deleteFile(bkt, fileName)
}

func (cli *StorageClient) uploadFile(bkt *storage.BucketHandle, fileName string) {
	if fileExists(bkt, fileName) {
		log.Println("file ", fileName, " already exists in cloud storage, not uploading.")
	}

	//open local file
	println(constants.LocalStorage + fileName)
	f, err := os.Open(constants.LocalStorage + fileName)
	i, _ := f.Stat()
	println("size of file: " + strconv.Itoa(int(i.Size())))

	if err != nil {
		log.Fatalf("failed to open local file before uploading: " + err.Error())
	}

	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	println("loading object: ")
	obj := bkt.Object(fileName)

	println("new writer: ")
	w := obj.NewWriter(ctx)
	if _, err = io.Copy(w, f); err != nil {
		log.Fatalf("io.Copy: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Writer.Close: %v", err)
	}
	println("file uploaded!")
}

func (cli *StorageClient) deleteFile(bkt *storage.BucketHandle, fileName string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	o := bkt.Object(fileName)
	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", fileName, err)
	}
	fmt.Println("blob deleted", fileName)
	return nil
}

func fileExists(bkt *storage.BucketHandle, fileName string) bool {
	query := &storage.Query{Prefix: fileName}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	println("checking if file exists")
	it := bkt.Objects(ctx, query)
	for {
		_, err := it.Next()
		if err == iterator.Done {
			println("file does not exist")
			return false
		}
		if err != nil {
			log.Fatal(err)
		}
		println("file exists")
		return true
	}
}

func (cli *StorageClient) DownloadSampleVideos() {
	bkt := cli.getSampleVideosBucketHandle()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//query := storage.Query{Prefix: token}
	objectIterator := bkt.Objects(ctx, &storage.Query{Prefix: ""})

	for {
		attrs, err := objectIterator.Next()
		if err == iterator.Done {
			println("iterator done")
			break
		}
		println("downloading sample")
		if err != nil {
			log.Fatal(err.Error())
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		rc, err := bkt.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Fatalf("DownloadSampleVideos: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}
		defer rc.Close()
		slurp, err := ioutil.ReadAll(rc)
		println("read " + strconv.Itoa(len(slurp)) + " bytes from downloaded file")
		if err != nil {
			log.Fatalf("DownloadSampleVideos: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}
		imagePath := constants.LocalStorage + attrs.Name
		println("Download imagePath: ", imagePath)
		f, err := os.Create(imagePath)
		if err != nil {
			log.Fatalf("DownloadSampleVideos, create file: %v", err)
		}
		_, err = f.Write(slurp)
		if err != nil {
			log.Fatalf("DownloadSampleVideos, write to file: %v", err)
		}

		f.Close()

	}
}

func (cli *StorageClient) DownloadUnconvertedPart(token string) {
	bkt := cli.getUnconvertedBucketHandle()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	rc, err := bkt.Object(token).NewReader(ctx)
	if err != nil {
		log.Fatalf("DownloadUnconvertedPart: unable to open file from bucket %q, file %q: %v", constants.UnconvertedVideosBucketName, token, err.Error())
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

func (cli *StorageClient) DownloadConvertedParts(token string) {
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
		println("downloading part: ", attrs.Name)
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		rc, err := bkt.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Printf("DownloadConvertedParts: unable to open file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}
		defer rc.Close()
		slurp, err := ioutil.ReadAll(rc)
		println("read " + strconv.Itoa(len(slurp)) + " bytes from downloaded file")
		if err != nil {
			log.Printf("DownloadConvertedParts: unable to read file from bucket %q, file %q: %v", constants.ConvertedVideosBucketName, attrs.Name, err)
			return
		}

		f, err := os.Create(constants.LocalStorage + attrs.Name + ".converted")
		if err != nil {
			log.Fatalf("DownloadConvertedParts, create file: %v", err)
		}
		_, err = f.Write(slurp)
		if err != nil {
			log.Fatalf("DownloadConvertedParts, write to file: %v", err)
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

func (cli *StorageClient) listBuckets() ([]string, error) {
	// projectID := "my-project-id"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var buckets []string
	it := client.Buckets(ctx, constants.ProjectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
		fmt.Println("Bucket: ", battrs.Name)
	}
	return buckets, nil
}

func (cli *StorageClient) DeleteConvertedParts(token string) {
	allFiles := cli.getConvertedVideos()
	for _, file := range allFiles {
		if strings.Split(file, "-")[0] == token {
			println("Deleting ", file, " from storage.")
			cli.DeleteConvertedPart(file)
		}
	}
}
func (cli *StorageClient) DeleteUnconvertedParts(token string) {
	allFiles := cli.getUnconvertedVideos()
	for _, file := range allFiles {
		if strings.Split(file, "-")[0] == token {
			println("Deleting ", file, " from storage.")
			cli.DeleteUnconvertedPart(file)
		}
	}
}

func (cli *StorageClient) DeleteAll() {
	allConvertedFiles := cli.getConvertedVideos()
	for _, file := range allConvertedFiles {
		println("Deleting ", file, " from storage.")
		cli.DeleteConvertedPart(file)
	}
	allUnconvertedFiles := cli.getUnconvertedVideos()
	for _, file := range allUnconvertedFiles {
		println("Deleting ", file, " from storage.")
		cli.DeleteUnconvertedPart(file)
	}
}
