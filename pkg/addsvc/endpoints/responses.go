package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
	"github.com/cage1016/gokitistiok8s/pkg/shared_package/responses"
)

var (
	_ httptransport.Headerer = (*SumResponse)(nil)

	_ httptransport.StatusCoder = (*SumResponse)(nil)

	_ httptransport.Headerer = (*ConcatResponse)(nil)

	_ httptransport.StatusCoder = (*ConcatResponse)(nil)
)

// SumResponse collects the response values for the Sum method.
type SumResponse struct {
	Rs  int64 `json:"rs"`
	Err error `json:"err,omitempty"`
}

func (r SumResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r SumResponse) Headers() http.Header {
	return http.Header{}
}

func (r SumResponse) Response() interface{} {
	return responses.DataRes{ApiVersion: service.Version, Data: r}
}

// ConcatResponse collects the response values for the Concat method.
type ConcatResponse struct {
	Rs  string `json:"rs"`
	Err error  `json:"err,omitempty"`
}

func (r ConcatResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r ConcatResponse) Headers() http.Header {
	return http.Header{}
}

func (r ConcatResponse) Response() interface{} {
	return responses.DataRes{ApiVersion: service.Version, Data: r}
}
