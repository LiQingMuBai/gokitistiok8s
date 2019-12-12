package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
)

type serviceTestSuite struct {
	svc service.AddsvcService
}

func (s *serviceTestSuite) SetupSubTest(_ *testing.T) func(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)
	s.svc = service.New(logger)
	return func(t *testing.T) {}
}

func setupServiceTestCase(t *testing.T) (serviceTestSuite, func(t *testing.T)) {
	return serviceTestSuite{}, func(t *testing.T) {

	}
}

func TestStubAddsvcService_Sum(t *testing.T) {
	s, teardownTestCase := setupServiceTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc string
		a, b int64
		want int64
	}{
		{
			desc: "sum 1 2",
			a:    1,
			b:    2,
			want: 3,
		},
		{
			desc: "sum 10 20",
			a:    10,
			b:    20,
			want: 30,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := s.svc.Sum(ctx, tc.a, tc.b)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			assert.Equal(t, tc.want, r, fmt.Sprintf("%s: expected res %d got %d", tc.desc, tc.want, r))
		})
	}
}

func TestStubAddsvcService_Concat(t *testing.T) {
	s, teardownTestCase := setupServiceTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc string
		a, b string
		want string
	}{
		{
			desc: "concat 1 2",
			a:    "1",
			b:    "2",
			want: "12",
		},
		{
			desc: "concat 10 20",
			a:    "10",
			b:    "20",
			want: "1020",
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := s.svc.Concat(ctx, tc.a, tc.b)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			assert.Equal(t, tc.want, r, fmt.Sprintf("%s: expected res %s got %s", tc.desc, tc.want, r))
		})
	}
}
