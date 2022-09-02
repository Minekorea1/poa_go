package watcher

import (
	context "context"
	"net"
	"os"
	"os/exec"
	"poa/log"
	"time"

	"google.golang.org/grpc"
)

var logger = log.NewLogger("watcher")

const GRPC_PORT = "8889"

type watcherServer struct {
	WatcherServiceServer

	MqttPublishTimestamp int64
	PoaArgs              []string
}

func (server *watcherServer) ImAlive(ctx context.Context, in *AliveInfo) (*AliveVoid, error) {
	server.MqttPublishTimestamp = in.MqttPublishTimestamp
	server.PoaArgs = in.PoaArgs
	return &AliveVoid{}, nil
}

func (server *watcherServer) GetAlive(ctx context.Context, in *AliveVoid) (*AliveInfo, error) {
	return &AliveInfo{MqttPublishTimestamp: server.MqttPublishTimestamp, PoaArgs: server.PoaArgs}, nil
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

func RunClient() *AliveInfo {
	conn, err := grpc.Dial("localhost:"+GRPC_PORT, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.LogfE("did not connect: %v", err)
	}
	defer conn.Close()

	c := NewWatcherServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var aliveInfo *AliveInfo
	aliveInfo, err = c.GetAlive(ctx, &AliveVoid{})
	if err != nil {
		logger.LogfE("could not request: %v", err)
	}

	logger.LogI(aliveInfo.GetMqttPublishTimestamp())
	// logger.LogI(aliveInfo.GetPoaArgs())

	return aliveInfo
}

var originalWD, _ = os.Getwd()

func StartProcess(args ...string) (*os.Process, error) {
	logger.LogW("StartProcess ", args)
	argv0, err := exec.LookPath(args[0])
	if err != nil {
		return nil, err
	}

	process, err := os.StartProcess(argv0, args, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   originalWD,
	})
	if err != nil {
		logger.LogE(err)
		return nil, err
	}
	return process, nil
}
