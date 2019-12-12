package endpoints

import (
	"math"

	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
)

type Request interface {
	validate() error
}

// SumRequest collects the request parameters for the Sum method.
type SumRequest struct {
	A int64 `json:"a"`
	B int64 `json:"b"`
}

func (r SumRequest) validate() error {
	if r.B > 0 {
		if r.A > math.MaxInt64-r.B {
			return service.ErrMalformedEntity
		}
	} else {
		if r.A < math.MinInt64-r.B {
			return service.ErrMalformedEntity
		}
	}
	return nil // TBA
}

// ConcatRequest collects the request parameters for the Concat method.
type ConcatRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}

func (r ConcatRequest) validate() error {
	return nil // TBA
}
