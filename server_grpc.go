package Vault

import (
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"golang.org/x/net/context"
)

type grpcServer struct {
	hash     grpctransport.Handler
	validate grpctransport.Handler
}

/*
	A gRPC Server in Go Kit

	we implement the methods of the interface, calling the ServeGRPC method on the appropriate handler. This method will
	actually serve requests by first decoding them, calling the appropriate endpoint function, getting the response, and
	encoding it and sending it back to the client who made the request
*/

func (s *grpcServer) Hash(ctx context.Context, r *pb.HashRequest) (*pb.HashResponse, error) {

	_, resp, err := s.hash.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.HashResponse), nil
}

func (s *grpcServer) Validate(ctx context.Context, r *pb.ValidateRequest) (*pb.ValidateResponse, error) {

	_, resp, err := s.validate.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ValidateResponse), nil
}

/*
	HashRequest objek pada PB
	hashRequest objek pada service.go

	This function is an EncodeRequestFunc function defined by Go kit, and it is used to translate our own hashRequest type
	into a protocol buffer type that can be used to communicate with the client
*/
func EncodeGRPCHashRequest(ctx context.Context, r interface{}) (interface{}, error) {

	req := r.(hashRequest)

	return &pb.HashRequest{Password: req.Password}, nil
}

/*
	We are going to do this for both encoding and decoding requests and responses for both hash and validate endpoints
*/

func DecodeGRPCHashRequest(ctx context.Context, r interface{}) (interface{}, error) {

	req := r.(*pb.HashRequest)

	return hashRequest{Password: req.Password}, nil
}

func EncodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {

	res := r.(hashResponse)

	return &pb.HashResponse{Hash: res.Hash, Err: res.Err}, nil
}

func DecodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {

	res := r.(*pb.HashResponse)

	return hashResponse{Hash: res.Hash, Err: res.Err}, nil
}

func EncodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {

	req := r.(validateRequest)

	return &pb.ValidateRequest{Password: req.Password, Hash: req.Hash}, nil
}

func DecodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {

	req := r.(*pb.ValidateRequest)

	return validateRequest{Password: req.Password, Hash: req.Hash}, nil
}

func EncodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {

	res := r.(validateResponse)

	return &pb.ValidateResponse{Valid: res.Valid}, nil
}

func DecodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {

	res := r.(*pb.ValidateResponse)

	return validateResponse{Valid: res.Valid}, nil
}

/*
	Like our HTTP server, we take in a base context and the actual Endpoints implementation that we are exposing via the
	gRPC server. We create and return a new instance of our grpcServer type, setting the handlers for both hash and validate
	by callinggrpctransport.NewServer. We use our endpoint.Endpoint functions for our service and tell the service which of
	our encoding/decoding functions to use for each case.
*/
func NewGRPCServer(ctx context.Context, endpoints Endpoints) pb.VaultServer {

	return &grpcServer{

		hash: grpctransport.NewServer(ctx,
			endpoints.HashEndpoint,
			DecodeGRPCHashRequest,
			EncodeGRPCHashResponse),

		validate: grpctransport.NewServer(ctx,
			endpoints.ValidateEndpoint,
			DecodeGRPCValidateRequest,
			EncodeGRPCValidateResponse),
	}
}
