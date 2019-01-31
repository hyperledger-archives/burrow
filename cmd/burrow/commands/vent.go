package commands

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hyperledger/burrow/vent/config"
	"github.com/hyperledger/burrow/vent/logger"
	"github.com/hyperledger/burrow/vent/service"
	"github.com/hyperledger/burrow/vent/sqlsol"
	"github.com/hyperledger/burrow/vent/types"
	cli "github.com/jawher/mow.cli"
)

func Vent(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		cfg := config.DefaultFlags()

		dbAdapterOpt := cmd.StringOpt("db-adapter", cfg.DBAdapter, "Database adapter, 'postgres' or 'sqlite' (if built with the sqlite tag) are supported")
		dbURLOpt := cmd.StringOpt("db-url", cfg.DBURL, "PostgreSQL database URL or SQLite db file path")
		dbSchemaOpt := cmd.StringOpt("db-schema", cfg.DBSchema, "PostgreSQL database schema (empty for SQLite)")
		grpcAddrOpt := cmd.StringOpt("grpc-addr", cfg.GRPCAddr, "Address to connect to the Hyperledger Burrow gRPC server")
		httpAddrOpt := cmd.StringOpt("http-addr", cfg.HTTPAddr, "Address to bind the HTTP server")
		logLevelOpt := cmd.StringOpt("log-level", cfg.LogLevel, "Logging level (error, warn, info, debug)")
		specFileOpt := cmd.StringOpt("spec-file", cfg.SpecFile, "SQLSol json specification file full path")
		abiFileOpt := cmd.StringOpt("abi-file", cfg.AbiFile, "Event Abi specification file full path")
		abiDirOpt := cmd.StringOpt("abi-dir", cfg.AbiDir, "Path of a folder to look for event Abi specification files")
		specDirOpt := cmd.StringOpt("spec-dir", cfg.SpecDir, "Path of a folder to look for SQLSol json specification files")
		dbBlockTxOpt := cmd.BoolOpt("db-block", cfg.DBBlockTx, "Create block & transaction tables and persist related data (true/false)")

		cmd.Before = func() {
			// Rather annoying boilerplate here... but there is no way to pass mow.cli a pointer for it to fill you value
			cfg.DBAdapter = *dbAdapterOpt
			cfg.DBURL = *dbURLOpt
			cfg.DBSchema = *dbSchemaOpt
			cfg.GRPCAddr = *grpcAddrOpt
			cfg.HTTPAddr = *httpAddrOpt
			cfg.LogLevel = *logLevelOpt
			cfg.SpecFile = *specFileOpt
			cfg.AbiFile = *abiFileOpt
			cfg.AbiDir = *abiDirOpt
			cfg.SpecDir = *specDirOpt
			cfg.DBBlockTx = *dbBlockTxOpt
		}

		cmd.Spec = ""

		cmd.Action = func() {
			log := logger.NewLogger(cfg.LogLevel)
			consumer := service.NewConsumer(cfg, log, make(chan types.EventData))
			server := service.NewServer(cfg, log, consumer)

			parser, err := sqlsol.SpecLoader(cfg.SpecDir, cfg.SpecFile, cfg.DBBlockTx)
			if err != nil {
				log.Error("err", err)
				os.Exit(1)
			}
			abiSpec, err := sqlsol.AbiLoader(cfg.AbiDir, cfg.AbiFile)
			if err != nil {
				log.Error("err", err)
				os.Exit(1)
			}

			var wg sync.WaitGroup

			// setup channel for termination signals
			ch := make(chan os.Signal)

			signal.Notify(ch, syscall.SIGTERM)
			signal.Notify(ch, syscall.SIGINT)

			// start the events consumer
			wg.Add(1)

			go func() {
				if err := consumer.Run(parser, abiSpec, true); err != nil {
					log.Error("err", err)
					os.Exit(1)
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
			os.Exit(0)
		}
	}
}
