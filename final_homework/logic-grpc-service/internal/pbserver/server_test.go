package pbserver

import (
	"context"
	"net"
	"testing"

	"smart-recruitment/logic-grpc-service/internal/app"
	logicpb "smart-recruitment/logic-grpc-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestProtobufGRPCRegisterAndListJobs(t *testing.T) {
	service := app.NewService()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	logicpb.RegisterLogicServiceServer(server, New(service))
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("serve protobuf grpc: %v", err)
		}
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial protobuf grpc: %v", err)
	}
	defer conn.Close()

	client := logicpb.NewLogicServiceClient(conn)
	hr, err := client.Register(context.Background(), &logicpb.AuthRequest{
		Role:     "hr",
		Email:    "hr@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register hr: %v", err)
	}

	created, err := client.CreateJob(context.Background(), &logicpb.CreateJobRequest{
		HrId:        hr.User.Id,
		Title:       "AI应用开发工程师",
		Description: "负责 AI 应用开发",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	jobs, err := client.ListJobs(context.Background(), &logicpb.Empty{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Jobs) != 1 || jobs.Jobs[0].Id != created.Id {
		t.Fatalf("unexpected jobs: %#v", jobs.Jobs)
	}
}
