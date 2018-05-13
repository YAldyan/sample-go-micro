package Vault

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
	"net/http"
)

/*
	An HTTP Server in Go Kit
*/

func NewHTTPServer(ctx context.Context, endpoints Endpoints) http.Handler {

	m := http.NewServeMux()
	m.Handle("/hash", httptransport.NewServer(ctx, endpoints.HashEndpoint, decodeHashRequest, encodeResponse))
	m.Handle("/validate", httptransport.NewServer(ctx, endpoints.ValidateEndpoint, decodeValidateRequest, encodeResponse))

	return m
}
