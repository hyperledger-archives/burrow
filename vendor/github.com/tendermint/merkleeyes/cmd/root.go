package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/merkleeyes/app"
	"github.com/tendermint/tmlibs/cli"
)

const (
	defaultLogLevel = "error"
	FlagLogLevel    = "log_level"
	// TODO: fix up these flag names when we do a minor release
	FlagDBName = "dbName"
	FlagDBType = "dbType"
)

var RootCmd = &cobra.Command{
	Use:   "merkleeyes",
	Short: "Merkleeyes server",
	Long: `Merkleeyes server and other tools

Including:
        - Start the Merkleeyes server
	- Benchmark to check the underlying performance of the databases.
	- Dump to list the full contents of any persistent go-merkle database.
	`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		level := viper.GetString(FlagLogLevel)
		err = app.SetLogLevel(level)
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			app.SetTraceLogger()
		}
		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().StringP(FlagDBType, "t", "goleveldb", "type of backing db")
	RootCmd.PersistentFlags().StringP(FlagDBName, "d", "", "database name")
	RootCmd.PersistentFlags().String(FlagLogLevel, defaultLogLevel, "Log level")
}
