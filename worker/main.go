package main

import (
	"context"
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"helloworld"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	apiKey := os.Getenv("TEMPORAL_CLOUD_API_KEY")
	if apiKey == "" {
		log.Fatal("TEMPORAL_CLOUD_API_KEY environment variable not set")
	}

	namespace := os.Getenv("TEMPORAL_CLOUD_NAMESPACE")
	if namespace == "" {
		log.Fatal("TEMPORAL_CLOUD_NAMESPACE environment variable not set")
	}

	address := os.Getenv("TEMPORAL_CLOUD_ADDRESS")
	if address == "" {
		log.Fatal("TEMPORAL_CLOUD_ADDRESS environment variable not set")
	}

	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		log.Fatal("TEMPORAL_TASK_QUEUE environment variable not set")
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},
			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", namespace),
							method,
							req,
							reply,
							cc,
							opts...,
						)
					},
				),
			},
		},
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
