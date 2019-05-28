package commands

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	cli "github.com/jawher/mow.cli"
)

// Vent consumes EVM events and commits to a DB
func Vent(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		cmd.Command("start", "Start the Vent consumer service",
			func(cmd *cli.Cmd) {
				cfg := config.DefaultVentConfig()

				dbAdapterOpt := cmd.StringOpt("db-adapter", cfg.DBAdapter, "Database adapter, 'postgres' or 'sqlite' (if built with the sqlite tag) are supported")
				dbURLOpt := cmd.StringOpt("db-url", cfg.DBURL, "PostgreSQL database URL or SQLite db file path")
				dbSchemaOpt := cmd.StringOpt("db-schema", cfg.DBSchema, "PostgreSQL database schema (empty for SQLite)")
				grpcAddrOpt := cmd.StringOpt("grpc-addr", cfg.GRPCAddr, "Address to connect to the Hyperledger Burrow gRPC server")
				httpAddrOpt := cmd.StringOpt("http-addr", cfg.HTTPAddr, "Address to bind the HTTP server")
				logLevelOpt := cmd.StringOpt("log-level", cfg.LogLevel, "Logging level (error, warn, info, debug)")
				abiFileOpt := cmd.StringsOpt("abi", cfg.AbiFileOrDirs, "EVM Contract ABI file or folder")
				specFileOrDirOpt := cmd.StringsOpt("spec", cfg.SpecFileOrDirs, "SQLSol specification file or folder")
				dbBlockTxOpt := cmd.BoolOpt("db-block", cfg.DBBlockTx, "Create block & transaction tables and persist related data (true/false)")

				announceEveryOpt := cmd.StringOpt("announce-every", "5s", "Announce vent status every period as a Go duration, e.g. 1ms, 3s, 1h")

				cmd.Before = func() {
					// Rather annoying boilerplate here... but there is no way to pass mow.cli a pointer for it to fill you value
					cfg.DBAdapter = *dbAdapterOpt
					cfg.DBURL = *dbURLOpt
					cfg.DBSchema = *dbSchemaOpt
					cfg.GRPCAddr = *grpcAddrOpt
					cfg.HTTPAddr = *httpAddrOpt
					cfg.LogLevel = *logLevelOpt
					cfg.AbiFileOrDirs = *abiFileOpt
					cfg.SpecFileOrDirs = *specFileOrDirOpt
					cfg.DBBlockTx = *dbBlockTxOpt

					if *announceEveryOpt != "" {
						var err error
						cfg.AnnounceEvery, err = time.ParseDuration(*announceEveryOpt)
						if err != nil {
							output.Fatalf("could not parse announce-every duration %s: %v", *announceEveryOpt, err)
						}
					}
				}

				cmd.Spec = "--spec=<spec file or dir> --abi=<abi file or dir> [--db-adapter] [--db-url] [--db-schema] " +
					"[--db-block] [--grpc-addr] [--http-addr] [--log-level] [--announce-every=<duration>]"

				cmd.Action = func() {
					log := logger.NewLogger(cfg.LogLevel)
					consumer := service.NewConsumer(cfg, log, make(chan types.EventData))
					server := service.NewServer(cfg, log, consumer)

					projection, err := sqlsol.SpecLoader(cfg.SpecFileOrDirs, cfg.DBBlockTx)
					if err != nil {
						output.Fatalf("Spec loader error: %v", err)
					}
					abiSpec, err := abi.LoadPath(cfg.AbiFileOrDirs...)
					if err != nil {
						output.Fatalf("ABI loader error: %v", err)
					}

					var wg sync.WaitGroup

					// setup channel for termination signals
					ch := make(chan os.Signal)

					signal.Notify(ch, syscall.SIGTERM)
					signal.Notify(ch, syscall.SIGINT)

					// start the events consumer
					wg.Add(1)

					go func() {
						if err := consumer.Run(projection, abiSpec, true); err != nil {
							output.Fatalf("Consumer execution error: %v", err)
						}

						wg.Done()
					}()

					// start the http server
					wg.Add(1)

					go func() {
						server.Run()
						wg.Done()
					}()

					// wait for a termination signal from the OS and
					// gracefully shutdown the events consumer and the http server
					go func() {
						<-ch
						consumer.Shutdown()
						server.Shutdown()
					}()

					// wait until the events consumer and the http server are done
					wg.Wait()
				}
			})

		cmd.Command("schema", "Print JSONSchema for spec file format to validate table specs",
			func(cmd *cli.Cmd) {
				cmd.Action = func() {
					output.Printf(source.JSONString(types.EventSpecSchema()))
				}
			})
	}
}
