package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"time"

	"github.com/howeyc/gopass"
	"github.com/hyperledger/burrow/config/deployment"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	cli "github.com/jawher/mow.cli"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

// Keys runs as either client or server
func Keys(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		proxyLoc := cmd.StringOpt("c proxy", "", "Location of the proxy server")
		keysDir := cmd.StringOpt("keys-dir", ".keys", "Directory where keys are stored")

		proxyClient := func(output Output) keys.KeysClient {
			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			conn, err := grpc.Dial(*proxyLoc, opts...)
			if err != nil {
				output.Fatalf("Failed to connect to proxy server: %v", err)
			}
			return keys.NewKeysClient(conn)
		}

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

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					resp, err := c.GenerateKey(ctx, &keys.GenRequest{Passphrase: password, CurveType: curve.String(), KeyName: *keyName})
					if err != nil {
						output.Fatalf("failed to generate key: %v", err)
					}

					fmt.Printf("%v\n", resp.Address)
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					key, err := ks.Gen(password, curve)
					if err != nil {
						output.Fatalf("failed to generate key: %v", err)
					}
					if keyName != nil {
						err = ks.AddName(*keyName, key.Address)
						if err != nil {
							output.Fatalf("failed to add name to key: %v", err)
						}
					}
					fmt.Printf("%v\n", key.Address)
				}
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

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					resp, err := c.Hash(ctx, &keys.HashRequest{Hashtype: *hashType, Message: message})
					if err != nil {
						output.Fatalf("failed to get public key: %v", err)
					}

					fmt.Printf("%v\n", resp.GetHash())
				} else {
					var hasher hash.Hash
					switch *hashType {
					case "ripemd160":
						hasher = ripemd160.New()
					case "sha256":
						hasher = sha256.New()
					// case "sha3":
					default:
						output.Fatalf("unknown hash type %v", *hashType)
					}

					hasher.Write(message)

					fmt.Printf("%v\n", hasher.Sum(nil))
				}
			}
		})

		cmd.Command("export", "Export a key to tendermint format", func(cmd *cli.Cmd) {
			keyName := cmd.StringOpt("name", "", "name of key to use")
			keyAddr := cmd.StringOpt("addr", "", "address of key to use")
			passphrase := cmd.StringOpt("passphrase", "", "passphrase for encrypted key")
			keyTemplate := cmd.StringOpt("t template", deployment.DefaultKeysExportFormat, "template for export key")

			cmd.Action = func() {
				ks := keys.NewKeyStore(*keysDir, true)
				var addr crypto.Address
				var err error

				if *keyAddr != "" {
					addr, err = crypto.AddressFromHexString(*keyAddr)
					if err != nil {
						output.Fatalf("address not in correct format: %v", err)
					}
				} else {
					addr, err = ks.GetName(*keyName)
					if err != nil {
						output.Fatalf("unable to get key by name: %v", err)
					}
				}

				k, err := ks.GetKey(*passphrase, addr)
				if err != nil {
					output.Fatalf("unable to get key by address %s: %v", addr, err)
				}

				key := deployment.Key{
					Name:       *keyName,
					CurveType:  k.CurveType.String(),
					Address:    k.Address,
					PublicKey:  k.PublicKey.PublicKey[:],
					PrivateKey: k.PrivateKey.PrivateKey[:],
				}

				str, err := key.Dump(*keyTemplate)
				if err != nil {
					output.Fatalf("failed to template key: %v", err)
				}

				fmt.Printf("%s\n", str)
			}
		})

		cmd.Command("import", "import <priv key> | /path/to/keyfile | <key json>", func(cmd *cli.Cmd) {
			curveType := cmd.StringOpt("t curvetype", "ed25519", "specify the curve type of key to create. Supports 'secp256k1' (ethereum),  'ed25519' (tendermint)")
			noPassword := cmd.BoolOpt("n no-password", false, "don't use a password for this key")
			key := cmd.StringArg("KEY", "", "private key, filename, or raw json")

			cmd.Action = func() {
				var password string
				if !*noPassword {
					fmt.Printf("Enter Password:")
					pwd, err := gopass.GetPasswdMasked()
					if err != nil {
						os.Exit(1)
					}
					password = string(pwd)
				}

				var privKeyBytes []byte
				fileContents, err := ioutil.ReadFile(*key)
				if err == nil {
					*key = string(fileContents)
				}

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if (*key)[:1] == "{" {
						resp, err := c.ImportJSON(ctx, &keys.ImportJSONRequest{JSON: *key})
						if err != nil {
							output.Fatalf("failed to import json key: %v", err)
						}

						fmt.Printf("%v\n", resp.Address)
					} else {
						privKeyBytes, err = hex.DecodeString(*key)
						if err != nil {
							output.Fatalf("failed to hex decode key: %s", *key)
						}
						resp, err := c.Import(ctx, &keys.ImportRequest{Passphrase: password, KeyBytes: privKeyBytes, CurveType: *curveType})
						if err != nil {
							output.Fatalf("failed to import json key: %v", err)
						}

						fmt.Printf("%v\n", resp.Address)
					}
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					if (*key)[:1] == "{" {
						addr, err := ks.LocalImportJSON(password, *key)
						if err != nil {
							output.Fatalf("failed to import json key: %v", err)
						}

						fmt.Printf("%v\n", addr)
					} else {
						privKeyBytes, err = hex.DecodeString(*key)
						if err != nil {
							output.Fatalf("failed to hex decode key: %s", *key)
						}

						ty, err := crypto.CurveTypeFromString(*curveType)
						if err != nil {
							output.Fatalf("unrecognised curve type: %v", err)
						}

						k, err := keys.NewKeyFromPriv(ty, privKeyBytes)
						if err != nil {
							output.Fatalf("failed to import private key: %v", err)
						}

						err = ks.StoreKey(password, k)
						if err != nil {
							output.Fatalf("failed to save key: %v", err)
						}

						fmt.Printf("%v\n", k.Address)
					}
				}
			}
		})

		cmd.Command("pub", "public key", func(cmd *cli.Cmd) {
			name := cmd.StringOpt("name", "", "name of key to use")
			keyAddr := cmd.StringOpt("addr", "", "address of key to use")

			cmd.Action = func() {
				var addr *crypto.Address
				var err error
				if keyAddr != nil {
					*addr, err = crypto.AddressFromHexString(*keyAddr)
					if err != nil {
						output.Fatalf("address `%s` not in correct format: %v", keyAddr, err)
					}
				}

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					resp, err := c.PublicKey(ctx, &keys.PubRequest{Name: *name, Address: addr})
					if err != nil {
						output.Fatalf("failed to get public key: %v", err)
					}

					fmt.Printf("%X\n", resp.GetPublicKey())
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					var addr crypto.Address
					var err error

					if keyAddr != nil {
						addr, err = crypto.AddressFromHexString(*keyAddr)
						if err != nil {
							output.Fatalf("address not in correct format: %v", err)
						}
					} else {
						addr, err = ks.GetName(*name)
						if err != nil {
							output.Fatalf("unable to get key by name: %v", err)
						}
					}

					key, err := ks.GetKey("", addr)
					if err != nil {
						output.Fatalf("unable to get key by address %s: %v", addr, err)
					}
					fmt.Printf("%X\n", key.GetPublicKey())
				}
			}
		})

		cmd.Command("sign", "sign <some data>", func(cmd *cli.Cmd) {
			name := cmd.StringOpt("name", "", "name of key to use")
			keyAddr := cmd.StringOpt("addr", "", "address of key to use")
			msg := cmd.StringArg("MSG", "", "message to sign")
			passphrase := cmd.StringOpt("passphrase", "", "passphrase for encrypted key")

			cmd.Action = func() {
				ks := keys.NewKeyStore(*keysDir, true)
				var addr crypto.Address
				var err error

				if keyAddr != nil {
					addr, err = crypto.AddressFromHexString(*keyAddr)
					if err != nil {
						output.Fatalf("address not in correct format: %v", err)
					}
				} else {
					addr, err = ks.GetName(*name)
					if err != nil {
						output.Fatalf("unable to get key by name: %v", err)
					}
				}
				message, err := hex.DecodeString(*msg)
				if err != nil {
					output.Fatalf("failed to hex decode message: %v", err)
				}

				key, err := ks.GetKey(*passphrase, addr)
				if err != nil {
					output.Fatalf("unable to get key by address %s: %v", addr, err)
				}

				sig, err := key.PrivateKey.Sign(message)
				if err != nil {
					output.Fatalf("unable to get key by address %s: %v", addr, err)
				}

				fmt.Printf("%X\n", sig.Signature)
			}
		})

		cmd.Command("verify", "verify <some data> <sig> <pubkey>", func(cmd *cli.Cmd) {
			curveTypeOpt := cmd.StringOpt("t curvetype", "ed25519", "specify the curve type of key to create. Supports 'secp256k1' (ethereum),  'ed25519' (tendermint)")

			msg := cmd.StringArg("MSG", "", "hash/message to check")
			sig := cmd.StringArg("SIG", "", "signature")
			pub := cmd.StringArg("PUBLIC", "", "public key")

			cmd.Action = func() {
				message, err := hex.DecodeString(*msg)
				if err != nil {
					output.Fatalf("failed to hex decode message: %v", err)
				}
				curveType, err := crypto.CurveTypeFromString(*curveTypeOpt)
				if err != nil {
					output.Fatalf("invalid curve type: %v", err)
				}

				signatureBytes, err := hex.DecodeString(*sig)
				if err != nil {
					output.Fatalf("failed to hex decode signature: %v", err)
				}

				signature, err := crypto.SignatureFromBytes(signatureBytes, curveType)
				if err != nil {
					output.Fatalf("could not form signature: %v", err)
				}

				publickey, err := hex.DecodeString(*pub)
				if err != nil {
					output.Fatalf("failed to hex decode publickey: %v", err)
				}

				pubkey, err := crypto.PublicKeyFromBytes(publickey, curveType)
				if err != nil {

				}
				err = pubkey.Verify(message, signature)

				if err != nil {
					output.Fatalf("failed to verify: %v", err)
				}
				output.Printf("true\n")
			}
		})

		cmd.Command("addname", "add key name to existing address", func(cmd *cli.Cmd) {
			keyname := cmd.StringArg("NAME", "", "name of key to use")
			keyaddr := cmd.StringArg("ADDR", "", "address of key to use")

			cmd.Action = func() {
				addr, err := crypto.AddressFromHexString(*keyaddr)
				if err != nil {
					output.Fatalf("address `%s` not in correct format: %v", keyaddr, err)
				}

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					_, err = c.AddName(ctx, &keys.AddNameRequest{Keyname: *keyname, Address: addr})
					if err != nil {
						output.Fatalf("failed to add name to addr: %v", err)
					}
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					err = ks.AddName(*keyname, addr)
					if err != nil {
						output.Fatalf("failed to add name to addr: %v", err)
					}
				}
			}
		})

		cmd.Command("list", "list keys", func(cmd *cli.Cmd) {
			name := cmd.StringOpt("name", "", "name or address of key to use")

			cmd.Action = func() {
				var list []*keys.KeyID
				var err error

				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					resp, err := c.List(ctx, &keys.ListRequest{KeyName: *name})
					if err != nil {
						output.Fatalf("failed to list key: %v", err)
					}
					list = resp.Key
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					list, err = ks.List(*name)
				}
				if list == nil {
					list = make([]*keys.KeyID, 0)
				}
				bs, err := json.MarshalIndent(list, "", "    ")
				if err != nil {
					output.Fatalf("failed to json encode keys: %v", err)
				}
				fmt.Printf("%s\n", string(bs))
			}
		})

		cmd.Command("rmname", "remove key name", func(cmd *cli.Cmd) {
			name := cmd.StringArg("NAME", "", "key to remove")

			cmd.Action = func() {
				if *proxyLoc != "" {
					c := proxyClient(output)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					_, err := c.RemoveName(ctx, &keys.RemoveNameRequest{KeyName: *name})
					if err != nil {
						output.Fatalf("failed to remove key: %v", err)
					}
				} else {
					ks := keys.NewKeyStore(*keysDir, true)
					err := ks.RmName(*name)
					if err != nil {
						output.Fatalf("failed to remove key name: %v", err)
					}
				}

			}
		})
	}
}
