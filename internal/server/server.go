// START: types
package server

// START: imports
import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	api "github.com/percybear/wal/api/v1"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// END: imports

// START: config_authorizer
type Config struct {
	CommitLog  CommitLog
	Authorizer Authorizer
}

const (
	objectWildcard = "*"
	produceAction  = "produce"
	consumeAction  = "consume"
)

// END: config_authorizer

var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	*Config
	api.UnimplementedLogServer
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

// END: types

// START: newgrpcserver_before_auth
// START: newgrpcserver
// func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, func(context.Context) error, error) {
func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// END: newgrpcserver_before_auth
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64(
					"grpc.time_ns",
					duration.Nanoseconds(),
				)
			},
		),
	}
	// END: logger

	// START: metrics_traces
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}
	// END: metrics_traces

	// START: grpc_opts

	opts = append(opts, grpc.StreamInterceptor(
		// opts = append(opts, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			// START_HIGHLIGHT
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			// END_HIGHLIGHT
			grpc_auth.StreamServerInterceptor(authenticate),
		)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		// START_HIGHLIGHT
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
		// END_HIGHLIGHT
		grpc_auth.UnaryServerInterceptor(authenticate),
	)),
		// START_HIGHLIGHT
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	// END_HIGHLIGHT
	)
	//END: grpc_opts

	// START: newgrpcserver_before_auth
	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	// api.RegisterLogServer(gsrv,
	// 	Produce:       srv.Produce,
	// 	Consume:       srv.Consume,
	// 	ConsumeStream: srv.ConsumeStream,
	// 	ProduceStream: srv.ProduceStream,
	// })
	api.RegisterLogServer(gsrv, srv)
	// grpc.ServiceRegistrar(gsrv).RegisterService(&api.Log_ServiceDesc, api.LogServer(srv))

	return gsrv, nil
}

// END: newgrpcserver
// END: newgrpcserver_before_auth

// // START: newgrpcserver_before_auth
// // START: newgrpcserver
// func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, func(context.Context) error, error) {
// 	// END: newgrpcserver_before_auth
// 	logger := zap.L().Named("server")
// 	zapOpts := []grpc_zap.Option{
// 		grpc_zap.WithDurationField(
// 			func(duration time.Duration) zapcore.Field {
// 				return zap.Int64(
// 					"grpc.time_ns",
// 					duration.Nanoseconds(),
// 				)
// 			},
// 		),
// 	}
// 	// END: logger

// 	// Set up OpenTelemetry.
// 	// Set up propagator.
// 	// Set up trace provider.
// 	// Set up meter provider.
// 	// Set up log provider.
// 	ctx := context.Background()
// 	otelShutdown, err := cfg.SetupOTelSDK(ctx)

// 	// START: metrics_traces
// 	// trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
// 	// err := view.Register(ocgrpc.DefaultServerViews...)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// END: metrics_traces

// 	// Set up OpenTelemetry exporter:w

// 	// Set up trace provider.
// 	// tp, err := config.newTraceProvider()
// 	// traceProvider, err := cfg.NewTraceProvider()
// 	// if err != nil {
// 	// 	handleErr(err)
// 	// 	return
// 	// }
// 	// shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
// 	// otel.SetTracerProvider(tracerProvider)
// 	// if err != nil {
// 	// 	handleErr(err)
// 	// 	return
// 	// }
// 	// shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
// 	// otel.SetTracerProvider(tp)

// 	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
// 	if err != nil {
// 		logger.Fatal("failed to initialize exporter:", zap.Error(err))
// 	}

// 	// Set up OpenTelemetry tracer provider
// 	// tp := trace.NewTracerProvider(
// 	// 	trace.WithBatcher(exporter),
// 	// )
// 	// defer func() { _ = traceProvider.Shutdown(context.Background()) }()

// 	// Set the global tracer provider
// 	// otel.SetTracerProvider(traceProvider)

// 	// Set up OpenTelemetry metric exporter
// 	// telemetryExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
// 	// if err != nil {
// 	// 	logger.Fatal("failed to initialize exporter:", zap.Error(err))
// 	// }

// 	// Set up OpenTelemetry meter provider
// 	// mp := metric.NewMeterProvider(
// 	// 	metric.WithReader(telemetryExporter),
// 	// )
// 	// defer func() { _ = mp.Shutdown(context.Background()) }()

// 	// Set the global meter provider
// 	// otel.SetMeterProvider(mp)

// 	// Wrap Zap logger with OpenTelemetry:w
// 	// otelLogger := otelzap.New(logger)
// 	// otelzap.ReplaceGlobals(otelLogger)

// 	opts = append(opts, grpc.StreamInterceptor(
// 		// opts = append(opts, grpc.StreamInterceptor(
// 		grpc_middleware.ChainStreamServer(
// 			// START_HIGHLIGHT
// 			grpc_ctxtags.StreamServerInterceptor(),
// 			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
// 			// END_HIGHLIGHT
// 			grpc_auth.StreamServerInterceptor(authenticate),
// 		)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
// 		// START_HIGHLIGHT
// 		grpc_ctxtags.UnaryServerInterceptor(),
// 		grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
// 		// END_HIGHLIGHT
// 		grpc_auth.UnaryServerInterceptor(authenticate),
// 	)),
// 	// START_HIGHLIGHT
// 	// grpc.StatsHandler(&ocgrpc.ServerHandler{}),
// 	// END_HIGHLIGHT
// 	)
// 	// START: newgrpcserver_before_auth
// 	gsrv := grpc.NewServer(opts...)
// 	srv, err := newgrpcServer(config)
// 	if err != nil {
// 		return nil, otelShutdown, err
// 	}
// 	// api.RegisterLogServer(gsrv,
// 	// 	Produce:       srv.Produce,
// 	// 	Consume:       srv.Consume,
// 	// 	ConsumeStream: srv.ConsumeStream,
// 	// 	ProduceStream: srv.ProduceStream,
// 	// })
// 	api.RegisterLogServer(gsrv, srv)
// 	// grpc.ServiceRegistrar(gsrv).RegisterService(&api.Log_ServiceDesc, api.LogServer(srv))

// 	return gsrv, otelShutdown, nil
// }

// // END: newgrpcserver
// // END: newgrpcserver_before_auth

// START: newTraceProvider

// func newTraceProvider() (*trace.TracerProvider, error) {
// 	traceExporter, err := stdouttrace.New(
// 		stdouttrace.WithPrettyPrint())
// 	if err != nil {
// 		return nil, err
// 	}

// 	traceProvider := trace.NewTracerProvider(
// 		trace.WithBatcher(traceExporter,
// 			// Default is 5s. Set to 1s for demonstrative purposes.
// 			trace.WithBatchTimeout(time.Second)),
// 	)
// 	return traceProvider, nil
// }

// END: newTraceProvider

// START: request_response
// START: produce_authorize
func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (
	*api.ProduceResponse, error) {
	// START_HIGHLIGHT
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		produceAction,
	); err != nil {
		return nil, err
	}
	// END_HIGHLIGHT
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

// END: produce_authorize

// START: consume_authorize
func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (
	*api.ConsumeResponse, error) {
	// START_HIGHLIGHT
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		consumeAction,
	); err != nil {
		return nil, err
	}
	// END_HIGHLIGHT
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{Record: record}, nil
}

// END: consume_authorize
// END: request_response

// START: stream
func (s *grpcServer) ProduceStream(
	stream api.Log_ProduceStreamServer,
) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

// START: consume_stream

func (s *grpcServer) ConsumeStream(
	req *api.ConsumeRequest,
	stream api.Log_ConsumeStreamServer,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

// END: consume_stream
// END: stream

// START: commitlog
type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

// END: commitlog

// START: authorizer
type Authorizer interface {
	Authorize(subject, object, action string) error
}

// END: authorizer

// START: authenticate
func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return ctx, status.New(
			codes.Unauthenticated,
			"no transport security being used",
		).Err()
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

func subject(ctx context.Context) string {
	testCtx := ctx
	anyValue := testCtx.Value(subjectContextKey{}).(string)
	return anyValue
}

type subjectContextKey struct{}

// END: authenticate
