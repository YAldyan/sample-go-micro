package Microservice


// Service provides password hashing capabilities.
type Service interface {

	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)

}

/*
	Part of designing micro-services is being careful about where state is stored. Even though you will implement the methods 
	of a service in a single file, with access to global variables, you should never use them to store the per-request or even 
	per-service state. It's important to remember that each service is likely to be running on many physical machines multiple
	times, each with no access to the others' global variables.
*/
type vaultService struct{}

// NewService makes a new Service.
func NewService() Service {
	return vaultService{}
}

// Ensure that you import the appropriate bcrypt package (try golang.org/x/crypto/bcrypt)

/*
	Note that the receiver in the Hash method is just (vaultService); we don't capture the variable because there is no way 
	we can store state on an empty struct
*/
func (vaultService) Hash(ctx context.Context, password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (vaultService) Validate(ctx context.Context, password, hash string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return false, nil
	}

	return true, nil
}


// Modeling Method Calls with Requests and Responses 

/*
	To model remote method calls, you essentially create a struct for the incoming arguments and a struct for the return 
	arguments.
*/

type hashRequest struct {
	Password string `json:"password"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err string `json:"err,omitempty"`
}


/*
	The signature for decodeHashRequest is dictated by Go kit because it will later use it to decode HTTP requests on our 
	behalf. In this function, we just use json.Decoder to unmarshal the JSON into our hashRequest type
*/
func decodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	var req hashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
i
	f err != nil {
		return nil, err
	}

	return req, nil
}


type validateRequest struct {
	Password string `json:"password"`
	Hash string `json:"hash"`
}

type validateResponse struct {
	Valid bool `json:"valid"`
	Err string `json:"err,omitempty"`
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	
	var req validateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	
	if err != nil {
		return nil, err
	}

	return req, nil
}

/*
	The final thing we need to do is encode the response. In this case, we can write a single method to encode both the 
	hashResponse and validateResponse objects.
*/
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}


/*
	Making Endpoints for Service Methods

	In order to turn our service methods into endpoint.Endpoint functions, we're going to write a function that handles the 
	incoming hashRequest, calls the Hash service method, and depending on the response, builds and returns an appropriate 
	hashResponse object

	This function takes Service as an argument, which means that we can generate an endpoint from any implementation 
	of our Service interface. We then use a type assertion to specify that the request argument should, in fact, be 
	of type hashRequest. We call the Hash method, passing in the context and Password, which we get from hashRequest. 
	If all is well, we build hashResponse with the value we got back from the Hash method and return it
*/

func MakeHashEndpoint(srv Service) endpoint.Endpoint { 

	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(hashRequest)
		v, err := srv.Hash(ctx, req.Password)

		if err != nil {
			return hashResponse{v, err.Error()}, nil
		}

		return hashResponse{v, ""}, nil
	}
}


func MakeValidateEndpoint(srv Service) endpoint.Endpoint {
	
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		
		req := request.(validateRequest)
		v, err := srv.Validate(ctx, req.Password, req.Hash)

		if err != nil {
			return validateResponse{false, err.Error()}, nil
		}

		return validateResponse{v, ""}, nil
	}
}


/*
	Wrapping Endpoints into a Service Implementation

	Implementation dari Service Struct
*/

type Endpoints struct {
	HashEndpoint endpoint.Endpoint
	ValidateEndpoint endpoint.Endpoint
}

func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {

	req := hashRequest{Password: password}

	resp, err := e.HashEndpoint(ctx, req)
	if err != nil {
		return "", err
	}

	hashResp := resp.(hashResponse)
	if hashResp.Err != "" {
		return "", errors.New(hashResp.Err)
	}

	return hashResp.Hash, nil
}

func (e Endpoints) Validate(ctx context.Context, password, hash string) (bool, error) {

	req := validateRequest{Password: password, Hash: hash}
	
	resp, err := e.ValidateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}

	validateResp := resp.(validateResponse)
	if validateResp.Err != "" {
		return false, errors.New(validateResp.Err)
	}

	return validateResp.Valid, nil
}