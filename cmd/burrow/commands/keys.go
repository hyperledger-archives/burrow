package commands

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"time"

	"github.com/howeyc/gopass"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/pbkeys"
	"github.com/jawher/mow.cli"
	"google.golang.org/grpc"
)

func grpcKeysClient(output Output) pbkeys.KeysClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(keys.KeyHost+":"+keys.KeyPort, opts...)
	if err != nil {
		output.Fatalf("Failed to connect to grpc server: %v", err)
	}
	return pbkeys.NewKeysClient(conn)
}

func Keys(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		if keysHost := os.Getenv("MONAX_KEYS_HOST"); keysHost != "" {
			keys.DefaultHost = keysHost
		}
		if keysPort := os.Getenv("MONAX_KEYS_PORT"); keysPort != "" {
			keys.DefaultPort = keysPort
		}

		keys.KeyHost = *cmd.StringOpt("host", keys.DefaultHost, "set the host for talking to the key daemon")

		keys.KeyPort = *cmd.StringOpt("port", keys.DefaultPort, "set the port for key daemon")

		cmd.Command("server", "run keys server", func(cmd *cli.Cmd) {
			keys.KeysDir = *cmd.StringOpt("dir", keys.DefaultDir, "specify the location of the directory containing key files")

			cmd.Action = func() {
				err := keys.StartStandAloneServer(keys.KeyHost, keys.KeyPort)
				if err != nil {
					output.Fatalf("Failed to start server: %v", err)
				}
			}
		})

		cmd.Command("gen", "Generates a key using (insert crypto pkgs used)", func(cmd *cli.Cmd) {
			noPassword := cmd.BoolOpt("n no-password", false, "don't use a password for this key")

			keyType := cmd.StringOpt("t curvetype", "ed25519", "specify the curve type of key to create. Supports 'secp256k1' (ethereum),  'ed25519' (tendermint)")

			keyName := cmd.StringOpt("name", "", "name of key to use")

			cmd.Action = func() {
				curve, err := crypto.CurveTypeFromString(*keyType)
				if err != nil {
					output.Fatalf("Unrecognised curve type %v", *keyType)
				}

				var password string
				if !*noPassword {
					fmt.Printf("Enter Password:")
					pwd, err := gopass.GetPasswdMasked()
					if err != nil {
						os.Exit(1)
					}
					password = string(pwd)
				}

				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.GenerateKey(ctx, &pbkeys.GenRequest{Passphrase: password, Curvetype: curve.String(), Keyname: *keyName})
				if err != nil {
					output.Fatalf("failed to generate key: %v", err)
				}

				fmt.Printf("%v\n", resp.GetAddress())
			}
		})

		cmd.Command("hash", "hash <some data>", func(cmd *cli.Cmd) {
			hashType := cmd.StringOpt("t type", keys.DefaultHashType, "specify the hash function to use")

			hexByte := cmd.BoolOpt("hex", false, "the input should be hex decoded to bytes first")

			msg := cmd.StringArg("MSG", "", "message to hash")

			cmd.Action = func() {
				var message []byte
				var err error
				if *hexByte {
					message, err = hex.DecodeString(*msg)
					if err != nil {
						output.Fatalf("failed to hex decode message: %v", err)
					}
				} else {
					message = []byte(*msg)
				}

				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.Hash(ctx, &pbkeys.HashRequest{Hashtype: *hashType, Message: message})
				if err != nil {
					output.Fatalf("failed to get public key: %v", err)
				}

				fmt.Printf("%v\n", resp.GetHash())
			}
		})

		cmd.Command("export", "Export a key to tendermint format", func(cmd *cli.Cmd) {
			keyName := cmd.StringOpt("name", "", "name of key to use")
			keyAddr := cmd.StringOpt("addr", "", "address of key to use")
			passphrase := cmd.StringOpt("passphrase", "", "passphrase for encrypted key")

			cmd.Action = func() {
				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.Export(ctx, &pbkeys.ExportRequest{Passphrase: *passphrase, Name: *keyName, Address: *keyAddr})
				if err != nil {
					output.Fatalf("failed to export key: %v", err)
				}

				fmt.Printf("%s\n", resp.GetExport())
			}
		})

		cmd.Command("import", "import <priv key> | /path/to/keyfile | <key json>", func(cmd *cli.Cmd) {
			curveType := cmd.StringOpt("t curvetype", "ed25519", "specify the curve type of key to create. Supports 'secp256k1' (ethereum),  'ed25519' (tendermint)")
			key := cmd.StringArg("KEY", "", "private key, filename, or raw json")

			cmd.Action = func() {
				var privKeyBytes []byte
				var err error
				if _, err := os.Stat(*key); err == nil {
					privKeyBytes, err = ioutil.ReadFile(*key)
					if err != nil {
						output.Fatalf("Failed to read file %s: %v", *key, err)
					}
				}

				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if (*key)[:1] == "{" {
					resp, err := c.ImportJSON(ctx, &pbkeys.ImportJSONRequest{JSON: *key})
					if err != nil {
						output.Fatalf("failed to import json key: %v", err)
					}

					fmt.Printf("%X\n", resp.GetAddress())
				} else {
					privKeyBytes, err = hex.DecodeString(*key)
					if err != nil {
						output.Fatalf("failed to hex decode key")
					}
					resp, err := c.Import(ctx, &pbkeys.ImportRequest{Keybytes: privKeyBytes, Curvetype: *curveType})
					if err != nil {
						output.Fatalf("failed to import json key: %v", err)
					}

					fmt.Printf("%X\n", resp.GetAddress())

				}
			}
		})

		cmd.Command("pub", "public key", func(cmd *cli.Cmd) {
			name := cmd.StringOpt("name", "", "name of key to use")
			addr := cmd.StringOpt("addr", "", "address of key to use")

			cmd.Action = func() {
				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.PublicKey(ctx, &pbkeys.PubRequest{Name: *name, Address: *addr})
				if err != nil {
					output.Fatalf("failed to get public key: %v", err)
				}

				fmt.Printf("%X\n", resp.GetPub())
			}
		})

		cmd.Command("sign", "sign <some data>", func(cmd *cli.Cmd) {
			addr := cmd.StringOpt("addr", "", "address of key to use")
			msg := cmd.StringArg("HASH", "", "hash to sign")
			passphrase := cmd.StringOpt("passphrase", "", "passphrase for encrypted key")

			cmd.Action = func() {
				message, err := hex.DecodeString(*msg)
				if err != nil {
					output.Fatalf("failed to hex decode message: %v", err)
				}

				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.Sign(ctx, &pbkeys.SignRequest{Passphrase: *passphrase, Address: *addr, Hash: message})
				if err != nil {
					output.Fatalf("failed to get public key: %v", err)
				}
				fmt.Printf("%X\n", resp.GetSignature())
			}
		})

		cmd.Command("verify", "verify <some data> <sig> <pubkey>", func(cmd *cli.Cmd) {
			curveType := cmd.StringOpt("t curvetype", "ed25519", "specify the curve type of key to create. Supports 'secp256k1' (ethereum),  'ed25519' (tendermint)")

			msg := cmd.StringArg("MSG", "", "hash/message to check")
			sig := cmd.StringArg("SIG", "", "signature")
			pub := cmd.StringArg("PUBLIC", "", "public key")

			cmd.Action = func() {
				message, err := hex.DecodeString(*msg)
				if err != nil {
					output.Fatalf("failed to hex decode message: %v", err)
				}

				signature, err := hex.DecodeString(*sig)
				if err != nil {
					output.Fatalf("failed to hex decode signature: %v", err)
				}

				publickey, err := hex.DecodeString(*pub)
				if err != nil {
					output.Fatalf("failed to hex decode publickey: %v", err)
				}

				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_, err = c.Verify(ctx, &pbkeys.VerifyRequest{Curvetype: *curveType, Pub: publickey, Signature: signature, Hash: message})
				if err != nil {
					output.Fatalf("failed to verify: %v", err)
				}
			}
		})

		cmd.Command("add", "add key by name or addr", func(cmd *cli.Cmd) {
			name := cmd.StringArg("name", "", "name of key to use")
			addr := cmd.StringArg("addr", "", "address of key to use")

			cmd.Action = func() {
				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_, err := c.Add(ctx, &pbkeys.AddRequest{Keyname: *name, Address: *addr})
				if err != nil {
					output.Fatalf("failed to add name to addr: %v", err)
				}
			}
		})

		cmd.Command("list", "list key names", func(cmd *cli.Cmd) {
			name := cmd.StringOpt("name", "", "name of key to use")

			cmd.Action = func() {
				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				resp, err := c.List(ctx, &pbkeys.Name{*name})
				if err != nil {
					output.Fatalf("failed to list key names: %v", err)
				}
				for _, k := range resp.Key {
					fmt.Printf("key: %v\n", k)
				}
			}
		})

		cmd.Command("rm", "rm key by name", func(cmd *cli.Cmd) {
			name := cmd.StringArg("NAME", "", "key to remove")

			cmd.Action = func() {
				c := grpcKeysClient(output)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_, err := c.Remove(ctx, &pbkeys.Name{*name})
				if err != nil {
					output.Fatalf("failed to remove key: %v", err)
				}
			}
		})
	}
}
