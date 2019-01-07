package cmd

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
	"github.com/spf13/cobra"
)

var ventCmd = &cobra.Command{
	Use:   "vent",
	Short: "Vent - an EVM event to SQL database mapping layer",
	Run:   runVentCmd,
}

var cfg = config.DefaultFlags()

func init() {
	ventCmd.Flags().StringVar(&cfg.DBAdapter, "db-adapter", cfg.DBAdapter, "Database adapter, 'postgres' or 'sqlite' are fully supported")
	ventCmd.Flags().StringVar(&cfg.DBURL, "db-url", cfg.DBURL, "PostgreSQL database URL or SQLite db file path")
	ventCmd.Flags().StringVar(&cfg.DBSchema, "db-schema", cfg.DBSchema, "PostgreSQL database schema (empty for SQLite)")
	ventCmd.Flags().StringVar(&cfg.GRPCAddr, "grpc-addr", cfg.GRPCAddr, "Address to connect to the Hyperledger Burrow gRPC server")
	ventCmd.Flags().StringVar(&cfg.HTTPAddr, "http-addr", cfg.HTTPAddr, "Address to bind the HTTP server")
	ventCmd.Flags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level (error, warn, info, debug)")
	ventCmd.Flags().StringVar(&cfg.SpecFile, "spec-file", cfg.SpecFile, "SQLSol json specification file full path")
	ventCmd.Flags().StringVar(&cfg.AbiFile, "abi-file", cfg.AbiFile, "Event Abi specification file full path")
	ventCmd.Flags().StringVar(&cfg.AbiDir, "abi-dir", cfg.AbiDir, "Path of a folder to look for event Abi specification files")
	ventCmd.Flags().StringVar(&cfg.SpecDir, "spec-dir", cfg.SpecDir, "Path of a folder to look for SQLSol json specification files")
	ventCmd.Flags().BoolVar(&cfg.DBBlockTx, "db-block", cfg.DBBlockTx, "Create block & transaction tables and persist related data (true/false)")
}

// Execute executes the vent command
func Execute() {
	if err := ventCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func runVentCmd(cmd *cobra.Command, args []string) {
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
