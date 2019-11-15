package main

import (
	"context"
	"fmt"
	registration "github.com/VineethReddy02/gRPC/registration-demo/protobuf"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	fmt.Println("Hello client here!")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()
	c := registration.NewRegistrationServiceClient(cc)
	Register(c) // Unary request/response
	BulkRegister(c) // Client side streaming
	GetAllRegisteredData(c) // Server side streaming
	RegisterOneByOne(c) // Bi-directional streaming
}

func Register(client registration.RegistrationServiceClient) {
	fmt.Println("### Performing Unary operation ###")
	request := &registration.RegisterRequest{
		Name:                 "abc",
		Email:                "abc",
	}
	response, err := client.Register(context.Background(),request)
	if err != nil {
		log.Fatalf("Unable to register %s",err)
	}
	fmt.Println("Successfully registered",response.RegistrationId)
}

func BulkRegister(client registration.RegistrationServiceClient) {
	fmt.Println("### Performing client side streaming ###")
	request := []*registration.RegisterRequest{
		{
			Name:  "Tom",
			Email: "tom@grpc.com",
		},
		{
			Name:  "Harvey",
			Email: "harvey@grpc.com",
		},
		{
			Name:  "Mike",
			Email: "mike@grpc.com",
		},
	}

	clientSender, err := client.RegisterBulk(context.Background())
	if err != nil {
		log.Fatalf("Unable to get sender client from bulk register %s",clientSender)
	}
	for _,req := range request {
		fmt.Println("Request sent to register: ",req.Name)
		err = clientSender.Send(req)
		if err != nil {
			log.Fatalf("Unable to send the register request for %s",err)
		}
	}
	result, err := clientSender.CloseAndRecv()
	if err != nil {
		log.Fatalf("Unable to close and receive end result %s",err)
	}

	fmt.Println("Successfully registered all the requests",result.BulkResponse)
}

func GetAllRegisteredData(request registration.RegistrationServiceClient) {
	fmt.Println("### Requesting for server side streaming ###")
	nothing := &registration.Nothing{}
	result,err := request.GetRegisteredData(context.Background(),nothing)
	if err != nil {
		log.Fatalf("Unable to get the receiver from get all data service")
	}
	for {
		res, err := result.Recv()
		if err == io.EOF {
			// We reached end of stream
			break
		}
		fmt.Println("Registered Id: ",res.RegistrationId)
	}
}

func RegisterOneByOne(request registration.RegistrationServiceClient) {
	fmt.Println("### Performing Bi-directional streaming ###")
	requests := []*registration.RegisterRequest{
		{
			Name:  "Tom",
			Email: "tom@grpc.com",
		},
		{
			Name:  "Harvey",
			Email: "harvey@grpc.com",
		},
		{
			Name:  "Mike",
			Email: "mike@grpc.com",
		},
	}

	client, err := request.RegisterMultipleRequests(context.Background())
	if err != nil {
		log.Fatalf("Unable recieve the client for registering multiple requets: %s",err)
	}

	waitc := make(chan struct{})

	go func(){
		for _, req := range requests {
			fmt.Println("Sending request to register: ",req.Name)
			err = client.Send(req)
			if err != nil {
				log.Print("Failed to send request of",req.Name)
			}
		}
		err = client.CloseSend()
		if err != nil {
			log.Fatalf("Unable to close and send stream for the client %v", err)
		}
	}()

	go func() {
		for {
			res, err := client.Recv()
			if err == io.EOF {
				break
			}
			fmt.Printf("Successfully registered: %s \n", res.RegistrationId)
		}
		close(waitc)
	}()

	<-waitc
}