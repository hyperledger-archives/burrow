package commands

import (
	"fmt"
	"net"

	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/proxy"
	cli "github.com/jawher/mow.cli"
	"google.golang.org/grpc"
)

// Proxy starts a proxy node
func Proxy(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainURLOpt := cmd.StringOpt("c chain", "127.0.0.1:10997", "chain to be used in IP:PORT format")
		keysDir := cmd.StringOpt("dir", ".keys", "specify the location of the directory containing key files")
		badPerm := cmd.BoolOpt("allow-bad-perm", false, "Allow unix key file permissions to be readable other than user")
		listenHostOpt := cmd.StringOpt("h listen-host", "localhost", "The host to listen on")
		listenPortOpt := cmd.IntOpt("p listen-port", 10998, "The port to listen on")

		cmd.Action = func() {
			server := keys.StandAloneServer(*keysDir, *badPerm)
			address := fmt.Sprintf("%s:%d", *listenHostOpt, *listenPortOpt)
			listener, err := net.Listen("tcp", address)
			if err != nil {
				output.Fatalf("Could not listen on %s: %v", address, err)
			}
			client, err := grpc.Dial(*chainURLOpt, grpc.WithInsecure())
			if err != nil {
				output.Fatalf("Error connecting to Burrow gRPC server at %s", chainURLOpt)
			}
			defer client.Close()

			proxy.New(client, server)

			err = server.Serve(listener)
			if err != nil {
				output.Fatalf("Keys server terminated with error: %v", err)
			}
		}
	}
}
