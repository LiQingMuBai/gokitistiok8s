package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokitistiok8s/pkg/foosvc/service"
	"github.com/cage1016/gokitistiok8s/pkg/shared_package/responses"
)

var (
	_ httptransport.Headerer = (*FooResponse)(nil)

	_ httptransport.StatusCoder = (*FooResponse)(nil)
)

// FooResponse collects the response values for the Foo method.
type FooResponse struct {
	Res string `json:"res"`
	Err error  `json:"err,omitempty"`
}

func (r FooResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r FooResponse) Headers() http.Header {
	return http.Header{}
}

func (r FooResponse) Response() interface{} {
	return responses.DataRes{ApiVersion: service.Version, Data: r}
}
