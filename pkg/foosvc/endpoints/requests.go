package endpoints

import "github.com/cage1016/gokitistiok8s/pkg/foosvc/service"

type Request interface {
	validate() error
}

// FooRequest collects the request parameters for the Foo method.
type FooRequest struct {
	S string `json:"s"`
}

func (r FooRequest) validate() error {
	if r.S == "" {
		return service.ErrMalformedEntity
	}
	return nil // TBA
}
