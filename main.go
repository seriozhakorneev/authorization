package main

import (
	"log"
	"net"

	"authorization/credentials"
	server "authorization/grpc"
	pb "authorization/grpc/authorize_proto"

	"google.golang.org/grpc"
)

func main() {
	creds, err := credentials.NewCredentials()
	if err != nil {
		log.Fatalln("credentials.NewCredentials: ", err)
	}

	port := "50051"

	s := grpc.NewServer()
	pb.RegisterAuthorizationServiceServer(
		s,
		&server.AuthorizationServer{Credentials: creds},
	)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("net.Listen: %v", err)
	}

	log.Println("Server listening on port:", port)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("s.Serve: %v", err)
	}

}
