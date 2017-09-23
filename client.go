package httpclient

import (
	"net/http"
	"time"

	"gopkg.in/eapache/go-resiliency.v1/breaker"
)

// Client is a standard utility interface that can be embedded in any larger interface.
// NOTE: if an API endpoint requires a different retry-policy or a different circuit-breaker configuration, etc.
// a new client will need to be created inside the handler and the new policies/configs can be specified via
// these methods. This will alter the behavior of the service Client's methods.
type Client interface {
	Do(*http.Request) (*http.Response, error)
	SetRetryPolicy(RetryPolicy)
	SetCircuitBreaker(Breaker)
	SetTransport(*http.Transport)
	GetStatus() bool
	SetStatus(bool)
}

// GenericClient ...
type GenericClient struct {
	name           string
	client         *http.Client
	retryPolicy    RetryPolicy
	circuitBreaker Breaker
	status         Status
}

// NewGenericClient ...
func NewGenericClient(name string, status Status) Client {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}

	genericClient := &GenericClient{
		name:           name,
		client:         &http.Client{Transport: tr},
		retryPolicy:    NewConstantRetryPolicy(100*time.Millisecond, 3),
		circuitBreaker: NewBreaker(DefaultBreakerConfig),
		status:         status,
	}

	genericClient.SetStatus(true)

	return genericClient
}

// Do ...
func (g *GenericClient) Do(req *http.Request) (*http.Response, error) {
	response := &http.Response{}
	b := g.circuitBreaker.CB
	backoffs := g.retryPolicy.Backoffs()
	retries := 0
	for {
		result := b.Run(func() error {
			resp, err := g.client.Do(req)
			response = resp
			if err != nil {
				return err
			}
			return nil
		})

		switch result {
		case nil:
			// Success
			g.SetStatus(true)
			return response, nil
		case breaker.ErrBreakerOpen:
			// Circuit-Breaker open
			g.SetStatus(false)
			return response, breaker.ErrBreakerOpen
		default:
			// Otherwise, retry
			if retries <= len(backoffs) && g.retryPolicy.Retry(req) {
				time.Sleep(backoffs[retries])
				retries++
			} else {
				g.SetStatus(false)
				return response, breaker.ErrBreakerOpen
			}
		}
	}

}

// SetRetryPolicy allows you to change the default retry-policy for the client
// NOTE: changing the retry-policy for a http client shared between go-routines will cause a data-race
func (g *GenericClient) SetRetryPolicy(policy RetryPolicy) {
	g.retryPolicy = policy
}

// SetCircuitBreaker allows you to change the default circuit-breaker for the client
// NOTE: changing the circuit-breaker for a http client shared between go-routines will cause a data-race
func (g *GenericClient) SetCircuitBreaker(breaker Breaker) {
	g.circuitBreaker = breaker
}

// SetTransport allows you to change the default http-transport for the client
// NOTE: changing the http-transport for a http client shared between go-routines will cause a data-race
func (g *GenericClient) SetTransport(transport *http.Transport) {
	g.client = &http.Client{Transport: transport}
}

// GetStatus allows you to read the current live status for the service.
func (g *GenericClient) GetStatus() bool {
	return g.status.Get(g.name)
}

// SetStatus allows you to set the current live status for the service.
func (g *GenericClient) SetStatus(status bool) {
	g.status.Set(g.name, status)
}
