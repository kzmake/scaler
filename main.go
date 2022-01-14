package main

import (
	"github.com/trusch/grpc-proxy/proxy"
	codec "github.com/trusch/grpc-proxy/proxy/codec"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	director := func(ctx context.Context, fullName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil, status.Errorf(codes.Internal, "failed to parse metadata: %v", err)
		}

		if _, found := md["grpc-accept-encoding"]; found {
			md.Delete("grpc-accept-encoding")
		}

		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())

		conn, err := grpc.DialContext(
			ctx,
			"localhost:50001",
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name())),
		)
		if err != nil {
			return ctx, nil, status.Errorf(codes.Internal, "failed to create connection: %v", err)
		}

		return outCtx, conn, nil
	}

	server := grpc.NewServer(
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	)

	proxy.RegisterService(server, director,
		"trusch.testproto.TestService",
		"PingEmpty", "Ping", "PingError", "PingList",
	)
}
