package ipc

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestServerClientSend(t *testing.T) {
	logger := zap.NewNop()

	handler := func(cmd Command) Response {
		return Response{Success: true, Message: "handled:" + cmd.Action, Data: cmd.Params}
	}

	srv, err := NewServer(handler, logger)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	defer srv.Stop()

	// start accept loop
	srv.Start()

	client := NewClient()

	resp, err := client.Send(Command{Action: "hello"})
	if err != nil {
		t.Fatalf("Send error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success, got: %#v", resp)
	}

	if resp.Message != "handled:hello" {
		t.Fatalf("unexpected message: %v", resp.Message)
	}

	// Ensure Stop cleans up socket file without error
	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestIsServerRunningAndTimeout(t *testing.T) {
	logger := zap.NewNop()

	handler := func(cmd Command) Response {
		return Response{Success: true}
	}

	srv, err := NewServer(handler, logger)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	defer srv.Stop()

	srv.Start()

	// give a little time for accept loop to be running (should be immediate)
	time.Sleep(10 * time.Millisecond)

	if !IsServerRunning() {
		t.Fatalf("expected server to be running")
	}

	// stop server and ensure client times out or errors
	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}

	// small timeout to provoke failure
	client := NewClient()
	_, err = client.SendWithTimeout(Command{Action: "ping"}, 100*time.Millisecond)
	if err == nil {
		t.Fatalf("expected error when server is stopped")
	}
}
