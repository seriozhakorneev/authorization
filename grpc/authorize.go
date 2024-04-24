package grpc

import (
	"context"
	"log"

	"authorization/authorize"
	"authorization/credentials"
	pb "authorization/grpc/authorize_proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationServer используется для реализации службы AuthorizationService
type AuthorizationServer struct {
	pb.UnimplementedAuthorizationServiceServer
	Credentials *credentials.Credentials
}

// GetAuthorizationData реализует AuthorizationService.GetAuthorizationData
func (s *AuthorizationServer) GetAuthorizationData(
	_ context.Context,
	_ *pb.AuthorizationDataRequest,
) (*pb.AuthorizationDataResponse, error) {

	cookies, err := authorize.Do(s.Credentials)
	if err != nil {
		log.Println("-------------Something went wrong: "+
			"\n----------------", err)
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	log.Println("----------Done\n" +
		"-----------------")
	return &pb.AuthorizationDataResponse{Cookies: cookies}, nil
}
