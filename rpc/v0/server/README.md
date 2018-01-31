# Server

This package contains classes for starting and running HTTP and websocket servers.

## Server interface

Servers implements the `Server` interface. A 'ServerConfig' and 'gin.Engine' object is supplied in the 'Start' method, so that they may set themselves up and set up the routes etc.

```
type Server interface {
	Start(*ServerConfig, *gin.Engine)
	Running() bool
	ShutDown()
}
```

The `Server` interface can be found in `server.go`.

## ServeProcess

The `ServeProcess` does the port binding and listening. You may attach any number of servers to the serve-process, and it will automatically call their 'Start' and 'ShutDown' methods when starting up and shutting down. You may also attach start and shutdown listeners to the `ServeProcess`.

The `ServeProcess` class can be found in `server.go`.

## WebSocketServer

The `WebSocketServer` is a template for servers that use websocket connections rather then HTTP. It will 

## Config

The config assumes that there is a default HTTP and Websocket server for RPC, and some other fields. See the main README.md for details.

While the system is generic (i.e. it does not care what a `Server` is or does), the configuration file is not. The reason is that the server is written specifically for burrow, and I do not want to manage generic config files (or perhaps even one per server).