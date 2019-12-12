package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	routertransport "github.com/cage1016/gokitistiok8s/pkg/router/transport"
)

const (
	defZipkinV2URL  = ""
	defServiceName  = "router"
	defLogLevel     = "error"
	defHTTPPort     = ""
	defGRPCPort     = ""
	defAddsvcURL    = ""
	defFoosvcURL    = ""
	defAuthnsvcURL  = ""
	defDomainsvcURL = ""
	defSecret       = "gokitistiok8s"

	envZipkinV2URL  = "QS_ZIPKIN_V2_URL"
	envServiceName  = "QS_ROUTER_SERVICE_NAME"
	envLogLevel     = "QS_ROUTER_LOG_LEVEL"
	envHTTPPort     = "QS_ROUTER_HTTP_PORT"
	envGRPCPort     = "QS_ROUTER_GRPC_PORT"
	envSecret       = "QS_ROUTER_SECRET"
	envAddsvcURL    = "QS_ADDSVC_URL"
	envFoosvcURL    = "QS_FOOSVC_URL"
	envAuthnsvcURL  = "QS_AUTHNSVC_URL"
	envDomainsvcURL = "QS_DOMAINSVC_URL"
)

const (
	routerAddsvc    = "add"
	routerFoosvc    = "foo"
	routerAuthnsvc  = "authn"
	routerDomainsvc = "domain"
)

// Env reads specified environment variable. If no value has been found,
// fallback is returned.
func env(key string, fallback string) (s0 string) {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type config struct {
	serviceName  string
	logLevel     string
	serviceHost  string
	httpPort     string
	grpcPort     string
	zipkinV2URL  string
	secret       string
	addsvcURL    string
	foosvcURL    string
	authnsvcURL  string
	domainsvcURL string
	routerMap    map[string]string
}

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = level.NewFilter(logger, level.AllowInfo())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	cfg := loadConfig(logger)
	logger = log.With(logger, "service", cfg.serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer := initOpentracing()
	zipkinTracer := initZipkin(cfg.serviceName, cfg.httpPort, cfg.zipkinV2URL, logger)

	logger.Log("routermap", fmt.Sprintf("%v", cfg.routerMap))

	hb := routertransport.NewHandlerBuilder()
	hb.AddHandler(routerAddsvc, routertransport.MakeAddSvcHandler(ctx, cfg.addsvcURL, tracer, zipkinTracer, logger))
	hb.AddHandler(routerFoosvc, routertransport.MakeFooSvcHandler(ctx, cfg.foosvcURL, tracer, zipkinTracer, logger))

	wg := &sync.WaitGroup{}
	go startHTTPServer(ctx, wg, hb.Router, cfg.httpPort, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	cancel()
	wg.Wait()

	level.Info(logger).Log("serviceName", cfg.serviceName, "terminated", "successful")
}

func loadConfig(logger log.Logger) (cfg config) {
	cfg.secret = env(envSecret, defSecret)
	cfg.serviceName = env(envServiceName, defServiceName)
	cfg.logLevel = env(envLogLevel, defLogLevel)
	cfg.httpPort = env(envHTTPPort, defHTTPPort)
	cfg.grpcPort = env(envGRPCPort, defGRPCPort)
	cfg.zipkinV2URL = env(envZipkinV2URL, defZipkinV2URL)
	cfg.addsvcURL = env(envAddsvcURL, defAddsvcURL)
	cfg.foosvcURL = env(envFoosvcURL, defFoosvcURL)
	cfg.authnsvcURL = env(envAuthnsvcURL, defAuthnsvcURL)
	cfg.domainsvcURL = env(envDomainsvcURL, defDomainsvcURL)

	cfg.routerMap = map[string]string{}
	cfg.routerMap[routerAddsvc] = cfg.addsvcURL
	cfg.routerMap[routerFoosvc] = cfg.foosvcURL
	cfg.routerMap[routerAuthnsvc] = cfg.authnsvcURL
	cfg.routerMap[routerDomainsvc] = cfg.domainsvcURL
	return
}

func initOpentracing() (tracer stdopentracing.Tracer) {
	return stdopentracing.GlobalTracer()
}

func initZipkin(serviceName, httpPort, zipkinV2URL string, logger log.Logger) (zipkinTracer *zipkin.Tracer) {
	var (
		err           error
		hostPort      = fmt.Sprintf("localhost:%s", httpPort)
		useNoopTracer = (zipkinV2URL == "")
		reporter      = zipkinhttp.NewReporter(zipkinV2URL)
	)
	zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
	zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer))
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
	if !useNoopTracer {
		logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinV2URL)
	}

	return
}

func startHTTPServer(ctx context.Context, wg *sync.WaitGroup, handler http.Handler, port string, logger log.Logger) {
	wg.Add(1)
	defer wg.Done()

	if port == "" {
		level.Error(logger).Log("protocol", "HTTP", "exposed", port, "err", "port is not assigned exist")
		return
	}

	p := fmt.Sprintf(":%s", port)
	// create a server
	srv := &http.Server{Addr: p, Handler: handler}
	level.Info(logger).Log("protocol", "HTTP", "exposed", port)
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			level.Info(logger).Log("Listen", err)
		}
	}()

	<-ctx.Done()

	// shut down gracefully, but wait no longer than 5 seconds before halting
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ignore error since it will be "Err shutting down server : context canceled"
	srv.Shutdown(shutdownCtx)

	level.Info(logger).Log("protocol", "HTTP", "Shutdown", "http server gracefully stopped")
}
