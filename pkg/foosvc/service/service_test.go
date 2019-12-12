package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/cage1016/gokitistiok8s/pkg/foosvc/mocks"
	"github.com/cage1016/gokitistiok8s/pkg/foosvc/service"
)

type serviceTestSuite struct {
	svc service.FoosvcService
}

func (s *serviceTestSuite) SetupSubTest(_ *testing.T) func(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	mockAddsvc := mocks.NewAddsvcMock()
	s.svc = service.New(mockAddsvc, logger)
	return func(t *testing.T) {}
}

func setupServiceTestCase(t *testing.T) (serviceTestSuite, func(t *testing.T)) {
	return serviceTestSuite{}, func(t *testing.T) {

	}
}

func TestStubFoosvcService_Foo(t *testing.T) {
	s, teardownTestCase := setupServiceTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc string
		foo  string
		want string
	}{
		{
			desc: "Foo a",
			foo:  "a",
			want: "a bar",
		},
		{
			desc: "Foo 124",
			foo:  "124",
			want: "124 bar",
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := s.svc.Foo(ctx, tc.foo)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			assert.Equal(t, tc.want, r, fmt.Sprintf("%s: expected res %s got %s", tc.desc, tc.want, r))
		})
	}
}
