// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/shultzenfegel/gojaeger/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/example/api"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	//tp, err := config.Init()
	tp, err := tracer.InitJaegerTracer("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	var conn *grpc.ClientConn
	conn, err = grpc.Dial(":7777", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer func() { _ = conn.Close() }()

	c := api.NewHelloServiceClient(conn)

	if err := callSayHello(c); err != nil {
		log.Fatal(err)
	}
	if err := callSayHelloClientStream(c); err != nil {
		log.Fatal(err)
	}
	if err := callSayHelloServerStream(c); err != nil {
		log.Fatal(err)
	}
	if err := callSayHelloBidiStream(c); err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)
}

func callSayHello(c api.HelloServiceClient) error {
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	response, err := c.SayHello(ctx, &api.HelloRequest{Greeting: "World"})
	if err != nil {
		return fmt.Errorf("calling SayHello: %w", err)
	}
	log.Printf("Response from server: %s", response.Reply)
	return nil
}

func callSayHelloClientStream(c api.HelloServiceClient) error {
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.SayHelloClientStream(ctx)
	if err != nil {
		return fmt.Errorf("opening SayHelloClientStream: %w", err)
	}

	for i := 0; i < 5; i++ {
		err := stream.Send(&api.HelloRequest{Greeting: "World"})

		time.Sleep(time.Duration(i*50) * time.Millisecond)

		if err != nil {
			return fmt.Errorf("sending to SayHelloClientStream: %w", err)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("closing SayHelloClientStream: %w", err)
	}

	log.Printf("Response from server: %s", response.Reply)
	return nil
}

func callSayHelloServerStream(c api.HelloServiceClient) error {
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.SayHelloServerStream(ctx, &api.HelloRequest{Greeting: "World"})
	if err != nil {
		return fmt.Errorf("opening SayHelloServerStream: %w", err)
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("receiving from SayHelloServerStream: %w", err)
		}

		log.Printf("Response from server: %s", response.Reply)
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func callSayHelloBidiStream(c api.HelloServiceClient) error {
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.SayHelloBidiStream(ctx)
	if err != nil {
		return fmt.Errorf("opening SayHelloBidiStream: %w", err)
	}

	serverClosed := make(chan struct{})
	clientClosed := make(chan struct{})

	go func() {
		for i := 0; i < 5; i++ {
			err := stream.Send(&api.HelloRequest{Greeting: "World"})

			if err != nil {
				// nolint: revive  // This acts as its own main func.
				log.Fatalf("Error when sending to SayHelloBidiStream: %s", err)
			}

			time.Sleep(50 * time.Millisecond)
		}

		err := stream.CloseSend()
		if err != nil {
			// nolint: revive  // This acts as its own main func.
			log.Fatalf("Error when closing SayHelloBidiStream: %s", err)
		}

		clientClosed <- struct{}{}
	}()

	go func() {
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				// nolint: revive  // This acts as its own main func.
				log.Fatalf("Error when receiving from SayHelloBidiStream: %s", err)
			}

			log.Printf("Response from server: %s", response.Reply)
			time.Sleep(50 * time.Millisecond)
		}

		serverClosed <- struct{}{}
	}()

	// Wait until client and server both closed the connection.
	<-clientClosed
	<-serverClosed
	return nil
}
