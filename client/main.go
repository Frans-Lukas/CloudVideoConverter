/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"bufio"
	"context"
	"github.com/Frans-Lukas/cloudvideoconverter/generated"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	upload()
	//download()

	/*for {
		helloWorld()
		time.Sleep(time.Second * 5)
	}*/
}

func upload() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := videoconverter.NewVideoConverterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := c.Upload(ctx)

	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token, err := c.RequestUploadToken(ctx2, &videoconverter.UploadTokenRequest{})
	if err != nil {
		println(err)
		return
	}

	req := videoconverter.Chunk{
		RequestType: &videoconverter.Chunk_Token{Token: token.Token},
	}

	stream.Send(&req)

	imagePath := "img.mp4"
	file, err := os.Open(imagePath)
	defer file.Close()

	v1, _ := os.Getwd()
	println(v1)

	if err != nil {
		println("cannot open file", err.Error())
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot read chunk to buffer: ", err)
		}

		req := &videoconverter.Chunk{
			RequestType: &videoconverter.Chunk_Content{Content: buffer[:n]},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.RetrievalToken)

}

func download() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := videoconverter.NewVideoConverterClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := videoconverter.DownloadRequest{Id: "test"}
	stream, err := c.Download(ctx, &request)

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Download: %v", err)
		}

		log.Println(data)
	}
}

/*func helloWorld() {
	println("trying to connect")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	println("connected")
	c := videoconverter.NewVideoConverterClient(conn)

	// Contact the server and print out its response.
}*/

/*func (c *videoconverter.VideoConverterClient) UploadFile(ctx context.Context, f string) (stats Stats, err error) {

	// Get a file handle for the file we
	// want to upload
	file, err = os.Open(f)

	// Open a stream-based connection with the
	// gRPC server
	stream, err := c.client.Upload(ctx)

	// Start timing the execution
	stats.StartedAt = time.Now()

	// Allocate a buffer with `chunkSize` as the capacity
	// and length (making a 0 array of the size of `chunkSize`)
	buf = make([]byte, c.chunkSize)
	for writing {
		// put as many bytes as `chunkSize` into the
		// buf array.
		n, err = file.Read(buf)

		// ... if `eof` --> `writing=false`...

		stream.Send(&messaging.Chunk{
			// because we might've read less than
			// `chunkSize` we want to only send up to
			// `n` (amount of bytes read).
			// note: slicing (`:n`) won't copy the
			// underlying data, so this as fast as taking
			// a "pointer" to the underlying storage.
			Content: buf[:n],
		})
	}

	// keep track of the end time so that we can take the elapsed
	// time later
	stats.FinishedAt = time.Now()

	// close
	status, err = stream.CloseAndRecv()
}*/
