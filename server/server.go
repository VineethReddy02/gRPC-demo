package main

import (
	"context"
	"fmt"
	registration "github.com/VineethReddy02/gRPC/registration-demo/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type server struct{}

var TotalRegisteredIds *registration.BulkRegisterResponse

func uniqueID() int {
	min := 1000
	max := 3000
	id := rand.Intn(max - min) + min
	return id
}

// Updates existing Total records in memory.
func updateTotalRegistrations(responses *registration.BulkRegisterResponse) {
	// This check makes sure memory is allocated on first run
	if TotalRegisteredIds == nil {
		TotalRegisteredIds = &registration.BulkRegisterResponse{}
	}
	// As long as the server is alive all the bulk registered are maintained in TotalRegisteredIds
	TotalRegisteredIds.BulkResponse = append(TotalRegisteredIds.BulkResponse, responses.BulkResponse...)
}

// This handles singe request/response flow
func (*server) Register(ctx context.Context, request *registration.RegisterRequest) (*registration.RegisterResponse, error) {
	fmt.Println("### Unary operation ###")
	id := uniqueID()
	response := &registration.RegisterResponse{
		RegistrationId:       strconv.Itoa(id),
	}
	fmt.Println("Registered: ",id)
	return response, nil
}

// This handles client side streaming to register multiple responses from client
func (*server) RegisterBulk(request registration.RegistrationService_RegisterBulkServer) error {
	fmt.Println("### Client side streaming ###")
	responses := &registration.BulkRegisterResponse{}
	for {
		req, err := request.Recv()
		id := uniqueID()
		if err == io.EOF {
			// maintains all registrations in memory
			updateTotalRegistrations(responses)
			return request.SendAndClose(responses)
		}
		response := &registration.RegisterResponse{
			RegistrationId:       strconv.Itoa(id),
		}
		fmt.Println("Received a request to register: ",req.Name)
		responses.BulkResponse = append(responses.BulkResponse,response)
		fmt.Println("Registered: ",req.Name)
	}
}

// Server side streaming sends all the registered data maintained in memory in multiple requests
func (*server) GetRegisteredData(request *registration.Nothing, dataServer registration.RegistrationService_GetRegisteredDataServer) error {
	fmt.Println("### Server side streaming ###")
	for _, data := range TotalRegisteredIds.BulkResponse {
		_ = dataServer.Send(data)
		fmt.Println("Server Streaming: Sending registered ID: ",data.RegistrationId)
		time.Sleep(2 * time.Second)
	}
	return nil
}

// Handles client side streaming by registering multiple requests from client
func(*server) RegisterMultipleRequests(request registration.RegistrationService_RegisterMultipleRequestsServer) error {
	fmt.Println("### Bi-directional streaming server/client ###")
	responses := &registration.BulkRegisterResponse{}
	for {
		req, err := request.Recv()
		if err == io.EOF {
			updateTotalRegistrations(responses)
			break
		} else if err != nil {
			log.Fatal("Unable to receive the message: ",err)
		}
		fmt.Println("Receieved request to register: ",req.Name)
		id := uniqueID()
		response := &registration.RegisterResponse{
			RegistrationId:       strconv.Itoa(id),
		}
		fmt.Println("Successfully registered: "+req.Name+" with "+ response.RegistrationId)
		_ = request.Send(response)
		time.Sleep(2*time.Second)
		responses.BulkResponse = append(responses.BulkResponse,response)
	}
	return nil
}

func main() {
	fmt.Println("Hello gRPC")

	lis, err := net.Listen("tcp","0.0.0.0:6666")
	if err != nil {
		log.Fatalf("failed to listen: %v",err)
	}
	s := grpc.NewServer()
	// Enable reflection to expose endpoints this server offers.
	reflection.Register(s)
	registration.RegisterRegistrationServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}