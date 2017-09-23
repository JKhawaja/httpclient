package httpclient

import "fmt"

// Status ...
type Status interface {
	Get(string) bool
	Set(string, bool) error
}

type setRequest struct {
	key   string
	value bool
	resp  chan bool
}

type getRequest struct {
	key  string
	resp chan bool
}

type servicesStatus struct {
	statuses    map[string]bool
	setRequests chan setRequest
	getRequests chan getRequest
}

// NewStatus ...
func NewStatus() Status {
	statuses := &servicesStatus{
		statuses:    make(map[string]bool),
		setRequests: make(chan setRequest),
		getRequests: make(chan getRequest),
	}

	go func(s *servicesStatus) {
		for {
			select {
			case get := <-s.getRequests:
				get.resp <- s.statuses[get.key]
			case set := <-s.setRequests:
				s.statuses[set.key] = set.value
				set.resp <- true
			}
		}
	}(statuses)

	return statuses
}

// Get ...
func (s *servicesStatus) Get(key string) bool {
	req := getRequest{
		key:  key,
		resp: make(chan bool),
	}

	s.getRequests <- req

	ret := <-req.resp

	return ret
}

// Set ...
func (s *servicesStatus) Set(key string, value bool) error {

	req := setRequest{
		key:   key,
		value: value,
		resp:  make(chan bool),
	}

	s.setRequests <- req

	ret := <-req.resp

	if ret == false {
		return fmt.Errorf("Could not set value.")
	}

	return nil
}
