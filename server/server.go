package server

import (
	"context"
	"net"
	"sync"
)

// Copies the signature from http.Server's graceful shutdown method
type Server interface {
	Shutdown(context.Context) error
}

type ShutdownFunc func(context.Context) error

func (sf ShutdownFunc) Shutdown(ctx context.Context) error {
	return sf(ctx)
}

type Launcher struct {
	Name   string
	Launch func() (Server, error)
}

type listenersServer struct {
	sync.Mutex
	listeners map[net.Listener]struct{}
}

// Providers a Server implementation from Listeners that are closed on shutdown
func FromListeners(listeners ...net.Listener) Server {
	lns := make(map[net.Listener]struct{}, len(listeners))
	for _, l := range listeners {
		lns[l] = struct{}{}
	}
	return &listenersServer{
		listeners: lns,
	}
}

func (ls *listenersServer) Shutdown(ctx context.Context) error {
	var err error
	for ln := range ls.listeners {
		if cerr := ln.Close(); cerr != nil && err == nil {
			err = cerr
		}
		delete(ls.listeners, ln)
	}
	return err
}
