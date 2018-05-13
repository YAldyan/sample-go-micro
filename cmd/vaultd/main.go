import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"Microservice/Vault"
	"Microservice/Vault/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	/*
		untuk membatasi maksimal akses ke endpoint tertentu
	*/
	ratelimitkit "github.com/go-kit/kit/ratelimit"
)

func main() {


	/*
		We use flags to allow the ops team to decide which endpoints we will listen on when
		exposing the service on the network, 
		but provide sensible defaults of 
				:8080 for the JSON/HTTP server and 
				:8081 for the gRPC server
	*/
	var (
			httpAddr = flag.String("http", ":8080", "http listen address")
			gRPCAddr = flag.String("grpc", ":8081", "gRPC listen address")
	)

	flag.Parse()

	/*
		We then create a new context using the context.Background() function, which returns a
		non-nil, empty context that has no cancelation or deadline specified and contains no 
		values, perfect for the base context of all of our services. Requests and middleware 
		are free to create new context objects from this one in order to add request-scoped 
		data or deadlines
	*/
	ctx := context.Background()

	srv := Vault.NewService()
	errChan := make(chan error)

	/*
		Buffered Channel

		Untuk secara real time menerima kesalahan/error program
	*/
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()


	/*
		Using Go Kit Endpoints
	*/
	hashEndpoint := vault.MakeHashEndpoint(srv)
	validateEndpoint := vault.MakeValidateEndpoint(srv)

	endpoints := vault.Endpoints{
									HashEndpoint: hashEndpoint,
									ValidateEndpoint: validateEndpoint,
								}
	/*
		endpoint dengan rate limiter

		hashEndpoint := vault.MakeHashEndpoint(srv){
			hashEndpoint = ratelimitkit.NewTokenBucketLimiter(rlbucket)(hashEndpoint)
		}

		validateEndpoint := vault.MakeValidateEndpoint(srv){
			validateEndpoint = ratelimitkit.NewTokenBucketLimiter(rlbucket)(validateEndpoint)
		}	
	
	*/							

	/*
		Running HTTP Server
	*/
	go func() {

		log.Println("http:", *httpAddr)
		handler := vault.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()


	/*
		Running gRPC Server
	*/
	go func() {
		
		listener, err := net.Listen("tcp", *gRPCAddr)
		if err != nil {
			errChan <- err
			return
		}

		log.Println("grpc:", *gRPCAddr)
		handler := vault.NewGRPCServer(ctx, endpoints)

		gRPCServer := grpc.NewServer()
		pb.RegisterVaultServer(gRPCServer, handler)

		errChan <- gRPCServer.Serve(listener)
	}()
}

/*
	We are going to use the NewTokenBucketLimiter middleware from Go kit's ratelimit package, and if we take a 
	look at the code, we'll see how it uses closures and returns functions to inject a call to the token bucket
	TakeAvailable method before passing execution to the next endpoint
*/
func NewTokenBucketLimiter(tb *ratelimit.Bucket) endpoint.Middleware {

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		
		return func(ctx context.Context, request interface{}) (interface{}, error) {

			if tb.TakeAvailable(1) == 0 {
				return nil, ErrLimited
			}

			return next(ctx, request)
		}
	}
}