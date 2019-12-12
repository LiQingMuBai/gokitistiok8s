package transports_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "github.com/cage1016/gokitistiok8s/pb/addsvc"
	"github.com/cage1016/gokitistiok8s/pkg/addsvc/endpoints"
	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
	"github.com/cage1016/gokitistiok8s/pkg/addsvc/transports"
)

const (
	port = 8181
)

func TestMain(m *testing.M) {
	startServer()
	code := m.Run()
	os.Exit(code)
}

func startServer() {
	logger := log.NewLogfmtLogger(os.Stderr)
	zkt, _ := zipkin.NewTracer(nil, zipkin.WithNoopTracer(true))
	tracer := opentracing.GlobalTracer()

	svc := service.New(logger)
	eps := endpoints.New(svc, logger, tracer, zkt)

	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := grpc.NewServer()
	pb.RegisterAddsvcServer(server, transports.MakeGRPCServer(eps, tracer, zkt, logger))
	go server.Serve(listener)
}

type grpcTransportsTestSuite struct {
	svc service.AddsvcService
}

func (s *grpcTransportsTestSuite) SetupSubTest(_ *testing.T) func(t *testing.T) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, _ := grpc.Dial(address, grpc.WithInsecure())

	logger := log.NewLogfmtLogger(os.Stderr)
	zkt, _ := zipkin.NewTracer(nil, zipkin.WithNoopTracer(true))
	tracer := opentracing.GlobalTracer()
	s.svc = transports.NewGRPCClient(conn, tracer, zkt, logger)

	return func(t *testing.T) {}
}

func setupGRPCTransportTestCase(t *testing.T) (grpcTransportsTestSuite, func(t *testing.T)) {
	return grpcTransportsTestSuite{}, func(t *testing.T) {

	}
}

func TestGRPCServer_Sum(t *testing.T) {
	s, teardownTestCase := setupGRPCTransportTestCase(t)
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

func TestGRPCServer_Concat(t *testing.T) {
	s, teardownTestCase := setupGRPCTransportTestCase(t)
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
