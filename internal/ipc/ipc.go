package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	// SocketName is the name of the Unix socket file
	SocketName = "neru.sock"

	// DefaultTimeout is the default timeout for IPC operations
	DefaultTimeout = 5 * time.Second

	// ConnectionTimeout is the timeout for establishing a connection
	ConnectionTimeout = 2 * time.Second
)

// Command represents an IPC command
type Command struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Response represents an IPC response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Server represents the IPC server
type Server struct {
	listener   net.Listener
	logger     *zap.Logger
	handler    CommandHandler
	socketPath string
}

// CommandHandler handles IPC commands
type CommandHandler func(cmd Command) Response

// GetSocketPath returns the path to the Unix socket
func GetSocketPath() string {
	tmpDir := os.TempDir()
	return filepath.Join(tmpDir, SocketName)
}

// NewServer creates a new IPC server
func NewServer(handler CommandHandler, logger *zap.Logger) (*Server, error) {
	socketPath := GetSocketPath()

	// Remove existing socket if it exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	logger.Info("IPC server created", zap.String("socket", socketPath))

	return &Server{
		listener:   listener,
		logger:     logger,
		handler:    handler,
		socketPath: socketPath,
	}, nil
}

// Start starts the IPC server
func (s *Server) Start() {
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				// If listener is closed, exit gracefully
				if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
					s.logger.Info("IPC server listener closed, stopping accept loop")
					return
				}
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}

			go s.handleConnection(conn)
		}
	}()
}

// handleConnection handles a single connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Error("Failed to close connection", zap.Error(err))
		}
	}()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var cmd Command
	if err := decoder.Decode(&cmd); err != nil {
		s.logger.Error("Failed to decode command", zap.Error(err))
		if encErr := encoder.Encode(Response{
			Success: false,
			Message: fmt.Sprintf("failed to decode command: %v", err),
		}); encErr != nil {
			s.logger.Error("Failed to encode error response", zap.Error(encErr))
		}
		return
	}

	s.logger.Info("Received command", zap.String("action", cmd.Action))

	response := s.handler(cmd)
	if err := encoder.Encode(response); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Stop stops the IPC server
func (s *Server) Stop() error {
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}

	// Clean up socket file
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Client represents an IPC client
type Client struct {
	socketPath string
}

// NewClient creates a new IPC client
func NewClient() *Client {
	return &Client{
		socketPath: GetSocketPath(),
	}
}

// Send sends a command to the IPC server with timeout
func (c *Client) Send(cmd Command) (Response, error) {
	return c.SendWithTimeout(cmd, DefaultTimeout)
}

// SendWithTimeout sends a command to the IPC server with a custom timeout
func (c *Client) SendWithTimeout(cmd Command, timeout time.Duration) (Response, error) {
	// Create a dialer with timeout
	dialer := net.Dialer{
		Timeout: ConnectionTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Response{}, fmt.Errorf("connection timeout: neru may be unresponsive")
		}
		return Response{}, fmt.Errorf("failed to connect to neru (is it running?): %w", err)
	}

	var closeErr error
	defer func() {
		if err := conn.Close(); err != nil && closeErr == nil {
			closeErr = fmt.Errorf("failed to close connection: %w", err)
		}
	}()

	// Set deadline for the entire operation
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return Response{}, fmt.Errorf("failed to set connection deadline: %w", err)
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(cmd); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Response{}, fmt.Errorf("send timeout: neru may be unresponsive")
		}
		err = fmt.Errorf("failed to send command: %w", err)
		if closeErr != nil {
			err = fmt.Errorf("%v (close error: %v)", err, closeErr)
		}
		return Response{}, err
	}

	var response Response
	if err := decoder.Decode(&response); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Response{}, fmt.Errorf("receive timeout: neru may be unresponsive")
		}
		err = fmt.Errorf("failed to receive response: %w", err)
		if closeErr != nil {
			err = fmt.Errorf("%v (close error: %v)", err, closeErr)
		}
		return Response{}, err
	}

	if closeErr != nil {
		return response, closeErr
	}
	return response, nil
}

// IsServerRunning checks if the IPC server is running
func IsServerRunning() bool {
	client := NewClient()
	_, err := client.Send(Command{Action: "ping"})
	return err == nil
}
