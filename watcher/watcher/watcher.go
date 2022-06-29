package watcher

import (
	context "context"
	"net"
	"poa/log"
	"time"

	"google.golang.org/grpc"
)

var logger = log.NewLogger("watcher")

const GRPC_PORT = "8889"

type watcherServer struct {
	WatcherServiceServer

	MqttPublishTimestamp int64
}

func (server *watcherServer) ImAlive(ctx context.Context, in *AliveTimestamp) (*AliveVoid, error) {
	server.MqttPublishTimestamp = in.MqttPublishTimestamp
	return &AliveVoid{}, nil
}

func (server *watcherServer) GetAlive(ctx context.Context, in *AliveVoid) (*AliveTimestamp, error) {
	return &AliveTimestamp{MqttPublishTimestamp: server.MqttPublishTimestamp}, nil
}

func RunServer() {
	logger.LogD("start watcher server")

	lis, err := net.Listen("tcp", ":"+GRPC_PORT)
	if err != nil {
		logger.LogfE("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	RegisterWatcherServiceServer(grpcServer, &watcherServer{})

	logger.LogfD("start gRPC server on %s port", GRPC_PORT)
	if err := grpcServer.Serve(lis); err != nil {
		logger.LogfE("failed to serve: %s", err)
	}
}

func RunClient() {
	conn, err := grpc.Dial("localhost:"+GRPC_PORT, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.LogfE("did not connect: %v", err)
	}
	defer conn.Close()

	c := NewWatcherServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var timestamp *AliveTimestamp
	timestamp, err = c.GetAlive(ctx, &AliveVoid{})
	if err != nil {
		logger.LogfE("could not request: %v", err)
	}

	logger.LogfD("Config: %v", timestamp)
}
