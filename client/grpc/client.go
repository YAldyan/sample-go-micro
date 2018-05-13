package grpc

import (
	"Go-Design-Pattern-For-Real-World/Microservice/Vault"
	"Go-Design-Pattern-For-Real-World/Microservice/Vault/pb"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

func New(conn *grpc.ClientConn) vault.Service {

	var hashEndpoint = grpctransport.NewClient(conn, "Vault", "Hash",
		vault.EncodeGRPCHashRequest,
		vault.DecodeGRPCHashResponse,
		pb.HashResponse{}).Endpoint()

	var validateEndpoint = grpctransport.NewClient(
		conn, "Vault", "Validate",
		vault.EncodeGRPCValidateRequest,
		vault.DecodeGRPCValidateResponse,
		pb.ValidateResponse{}).Endpoint()

	return vault.Endpoints{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}
}
