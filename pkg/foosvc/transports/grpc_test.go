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

	pb "github.com/cage1016/gokitistiok8s/pb/foosvc"
	"github.com/cage1016/gokitistiok8s/pkg/foosvc/endpoints"
	"github.com/cage1016/gokitistiok8s/pkg/foosvc/mocks"
	"github.com/cage1016/gokitistiok8s/pkg/foosvc/service"
	"github.com/cage1016/gokitistiok8s/pkg/foosvc/transports"
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

	mockAddsvc := mocks.NewAddsvcMock()
	svc := service.New(mockAddsvc, logger)
	eps := endpoints.New(svc, logger, tracer, zkt)

	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := grpc.NewServer()
	pb.RegisterFoosvcServer(server, transports.MakeGRPCServer(eps, tracer, zkt, logger))
	go server.Serve(listener)
}

type grpcTransportsTestSuite struct {
	svc service.FoosvcService
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

func TestGrpcServer_Foo(t *testing.T) {
	s, teardownTestCase := setupGRPCTransportTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		desc string
		s    string
		want string
	}{
		{
			desc: "foo 123",
			s:    "123",
			want: "123 bar",
		},
		{
			desc: "foo abc",
			s:    "abc",
			want: "abc bar",
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			teardownSubTest := s.SetupSubTest(t)
			defer teardownSubTest(t)

			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := s.svc.Foo(ctx, tc.s, )
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			assert.Equal(t, tc.want, r, fmt.Sprintf("%s: expected res %s got %s", tc.desc, tc.want, r))
		})
	}
}
