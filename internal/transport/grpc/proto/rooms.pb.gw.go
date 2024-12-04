// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: rooms/rooms.proto

/*
Package proto is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package proto

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_RoomsService_PingPong_0(ctx context.Context, marshaler runtime.Marshaler, client RoomsServiceClient, req *http.Request, pathParams map[string]string) (RoomsService_PingPongClient, runtime.ServerMetadata, chan error, error) {
	var metadata runtime.ServerMetadata
	errChan := make(chan error, 1)
	stream, err := client.PingPong(ctx)
	if err != nil {
		grpclog.Errorf("Failed to start streaming: %v", err)
		close(errChan)
		return nil, metadata, errChan, err
	}
	dec := marshaler.NewDecoder(req.Body)
	handleSend := func() error {
		var protoReq Ping
		err := dec.Decode(&protoReq)
		if err == io.EOF {
			return err
		}
		if err != nil {
			grpclog.Errorf("Failed to decode request: %v", err)
			return status.Errorf(codes.InvalidArgument, "Failed to decode request: %v", err)
		}
		if err := stream.Send(&protoReq); err != nil {
			grpclog.Errorf("Failed to send request: %v", err)
			return err
		}
		return nil
	}
	go func() {
		defer close(errChan)
		for {
			if err := handleSend(); err != nil {
				errChan <- err
				break
			}
		}
		if err := stream.CloseSend(); err != nil {
			grpclog.Errorf("Failed to terminate client stream: %v", err)
		}
	}()
	header, err := stream.Header()
	if err != nil {
		grpclog.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, errChan, err
	}
	metadata.HeaderMD = header
	return stream, metadata, errChan, nil
}

// RegisterRoomsServiceHandlerServer registers the http handlers for service RoomsService to "mux".
// UnaryRPC     :call RoomsServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterRoomsServiceHandlerFromEndpoint instead.
// GRPC interceptors will not work for this type of registration. To use interceptors, you must use the "runtime.WithMiddlewares" option in the "runtime.NewServeMux" call.
func RegisterRoomsServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server RoomsServiceServer) error {

	mux.Handle("GET", pattern_RoomsService_PingPong_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		err := status.Error(codes.Unimplemented, "streaming calls are not yet supported in the in-process transport")
		_, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
		return
	})

	return nil
}

// RegisterRoomsServiceHandlerFromEndpoint is same as RegisterRoomsServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterRoomsServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Errorf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterRoomsServiceHandler(ctx, mux, conn)
}

// RegisterRoomsServiceHandler registers the http handlers for service RoomsService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterRoomsServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterRoomsServiceHandlerClient(ctx, mux, NewRoomsServiceClient(conn))
}

// RegisterRoomsServiceHandlerClient registers the http handlers for service RoomsService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "RoomsServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "RoomsServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "RoomsServiceClient" to call the correct interceptors. This client ignores the HTTP middlewares.
func RegisterRoomsServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client RoomsServiceClient) error {

	mux.Handle("GET", pattern_RoomsService_PingPong_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/proto.RoomsService/PingPong", runtime.WithHTTPPathPattern("/ping-pong"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		resp, md, reqErrChan, err := request_RoomsService_PingPong_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}
		go func() {
			for err := range reqErrChan {
				if err != nil && err != io.EOF {
					runtime.HTTPStreamError(annotatedContext, mux, outboundMarshaler, w, req, err)
				}
			}
		}()

		forward_RoomsService_PingPong_0(annotatedContext, mux, outboundMarshaler, w, req, func() (proto.Message, error) { return resp.Recv() }, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_RoomsService_PingPong_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"ping-pong"}, ""))
)

var (
	forward_RoomsService_PingPong_0 = runtime.ForwardResponseStream
)
