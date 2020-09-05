package main

import (
	"flag"
	"fmt"
	//"net"
	"net/http"
	"os"
	"os/signal"
	//"strings"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/microservices-demo/payment"

	"github.com/opentracing/opentracing-go"
    "github.com/signalfx/signalfx-go-tracing/ddtrace/opentracer"
    sfxtracer "github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"

	"golang.org/x/net/context"
)

func main() {
	var (
		port          = flag.String("port", "8080", "Port to bind HTTP listener")
		declineAmount = flag.Float64("decline", 105, "Decline payments over certain amount")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	serviceName := os.Getenv("SIGNALFX_SERVICE_NAME")
	endpointURL := os.Getenv("SIGNALFX_ENDPOINT_URL")
	environment := os.Getenv("SIGNALFX_ENVIRONMENT")
	if environment != "" {
		sfxtracer.Start(
			sfxtracer.WithServiceName(serviceName),
			sfxtracer.WithZipkin(serviceName, endpointURL, ""),
			sfxtracer.WithGlobalTag("environment", environment))
	} else {
		sfxtracer.Start(
			sfxtracer.WithServiceName(serviceName),
			sfxtracer.WithZipkin(serviceName, endpointURL, ""))
	}
	defer sfxtracer.Stop()
	tracer := opentracer.New()
	opentracing.SetGlobalTracer(tracer)
	
	// Mechanical stuff.
	errc := make(chan error)
	ctx := context.Background()

	handler, logger := payment.WireUp(ctx, float32(*declineAmount), tracer, serviceName)

	// Create and launch the HTTP server.
	go func() {
		logger.Log("transport", "HTTP", "port", *port)
		errc <- http.ListenAndServe(":"+*port, handler)
	}()

	// Capture interrupts.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("exit", <-errc)
}
