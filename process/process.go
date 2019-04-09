package process

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// Copies the signature from http.Server's graceful shutdown method
type Process interface {
	Shutdown(context.Context) error
}

type ShutdownFunc func(context.Context) error

func (sf ShutdownFunc) Shutdown(ctx context.Context) error {
	return sf(ctx)
}

type Launcher struct {
	Name    string
	Enabled bool
	Launch  func() (Process, error)
}

func ListenerFromAddress(listenAddress string) (net.Listener, error) {
	const errHeader = "ListenerFromAddress():"
	parts := strings.Split(listenAddress, "://")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%s expects a fully qualified listen address like 'tcp://localhost:12345' but got '%s'",
			errHeader, listenAddress)
	}
	scheme := parts[0]
	address := parts[1]
	switch scheme {
	case "unix", "tcp":
	case "":
		scheme = "tcp"
	default:
		return nil, fmt.Errorf("%s did not recognise protocol %s in address '%s'", errHeader, scheme, listenAddress)
	}
	listener, err := net.Listen(scheme, address)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}
	return listener, nil
}
