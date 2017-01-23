package main

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"fmt"
	"os"
)

func main() {
	hellCmd := &cobra.Command{
		Use:   "hell",
		Short: "Hell makes the most of it being warm",
		Long:  "",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	var baseGlideLockFile, depGlideLockFile string
	dependCmd := &cobra.Command{
		Use:   "lock-merge",
		Short: "Merge glide.lock files together",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := ioutil.ReadFile(baseGlideLockFile)
			if err != nil {
				fmt.Printf("Could not read file: %s\n", err)
				os.Exit(1)
			}
			err = yaml.Unmarshal(bytes, &m)
			bytes, err = ioutil.ReadFile(depGlideLockFile)
			if err != nil {
				fmt.Printf("Could not read file: %s\n", err)
				os.Exit(1)
			}
			m := make(map[interface{}]interface{})
		},
	}
	dependCmd.PersistentFlags().StringVarP(&baseGlideLockFile, "base", "b", "", "")
	dependCmd.PersistentFlags().StringVarP(&depGlideLockFile, "dep", "d", "", "")
	hellCmd.AddCommand(dependCmd)
	dependCmd.Execute()
}
