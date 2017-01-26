package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-db/util/hell"

	"github.com/Masterminds/glide/action"
	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"
	"github.com/spf13/cobra"
)

func main() {
	hellCmd := &cobra.Command{
		Use:   "hell",
		Short: "Hell makes the most of it being warm",
		Long:  "",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	// Lock merge command
	var baseGlideLockFile, depGlideLockFile string
	lockMergeCmd := &cobra.Command{
		Use:   "lock-merge",
		Short: "Merge glide.lock files together",
		Long: "This command merges two glide.lock files into a single one by copying all dependencies " +
			"from a base glide.lock and an override glide.lock to an output glide.lock with dependencies " +
			"from override taking precedence over those from base.",
		Run: func(cmd *cobra.Command, args []string) {
			baseLockFile, err := cfg.ReadLockFile(baseGlideLockFile)
			if err != nil {
				fmt.Printf("Could not read file: %s\n", err)
				os.Exit(1)
			}
			overrideLockFile, err := cfg.ReadLockFile(depGlideLockFile)
			if err != nil {
				fmt.Printf("Could not read file: %s\n", err)
				os.Exit(1)
			}
			mergedLockFile, err := hell.MergeGlideLockFiles(baseLockFile, overrideLockFile)
			if err != nil {
				fmt.Printf("Could not merge lock files: %s\n", err)
				os.Exit(1)
			}
			mergedBytes, err := mergedLockFile.Marshal()
			if err != nil {
				fmt.Printf("Could not marshal lock file: %s\n", err)
				os.Exit(1)
			}
			os.Stdout.Write(mergedBytes)
		},
	}
	lockMergeCmd.PersistentFlags().StringVarP(&baseGlideLockFile, "base", "b", "", "base lock file")
	lockMergeCmd.PersistentFlags().StringVarP(&depGlideLockFile, "override", "o", "", "override lock file")

	// Lock update
	interactive := false
	getTransitiveCmd := &cobra.Command{
		Use:   "get",
		Short: "gets a remote dependency to this project along with its transtive dependencies.",
		Long: "Gets a remote dependency and its transitive dependencies by adding the remote " +
			"depednency to this project's glide.yaml and merging the remote dependency's " +
			"glide.lock into this project's glide.lock",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				msg.Die("%s requires a single argument of the remote dependency\n", cmd.Name())
			}
			rootPackage, _ := util.NormalizeName(args[0])
			// Add dependency to glide
			action.Get(args, repo.NewInstaller(), false, true, false, !interactive, false)
			// Now hunt down the repo cache
			dep := action.EnsureConfig().Imports.Get(rootPackage)
			key, err := cache.Key(dep.Remote())
			if err != nil {
				msg.Die("%s requires a single argument of the remote dependency\n", cmd.Name())
			}
			cacheDir := filepath.Join(cache.Location(), "src", key)

			if path.HasLock(cacheDir) {
				msg.Info("Found dependency lock file, merging into project lock file")
				lockPath := filepath.Join(".", path.LockFile)
				baseLockFile, err := cfg.ReadLockFile(lockPath)
				if err != nil {
					msg.Die("Could not read base lock file: %s", err)
				}
				overrideLockFile, err := cfg.ReadLockFile(filepath.Join(cacheDir, path.LockFile))
				if err != nil {
					msg.Die("Could not read dependency lock file: %s", err)
				}
				mergedLockFile, err := hell.MergeGlideLockFiles(baseLockFile, overrideLockFile)
				if err != nil {
					msg.Die("Could not merge lock files: %s\n", err)
				}
				err = mergedLockFile.WriteFile(lockPath)
				if err != nil {
					msg.Die("Could not write merged lock file: %s", err)
				}
			} else {
				msg.Info("Did not find dependency lock file, so nothing merged intoo project lock file")
			}
		},
	}

	getTransitiveCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false,
		"set dependency verion interactively")

	hellCmd.AddCommand(lockMergeCmd)
	hellCmd.AddCommand(getTransitiveCmd)
	lockMergeCmd.Execute()
}
