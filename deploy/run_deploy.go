package pkgs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/deploy/loader"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
)

type playbookWork struct {
	playbook string
	jobNo    int
}

type playbookResult struct {
	jobNo    int
	log      bytes.Buffer
	err      error
	duration time.Duration
}

func worker(playbooks <-chan playbookWork, results chan<- playbookResult, args *def.DeployArgs, logger *logging.Logger) {

	client := def.NewClient(args.Chain, args.KeysService, args.MempoolSign, time.Duration(args.Timeout)*time.Second)

	for playbook := range playbooks {
		doWork := func(work playbookWork) (logBuf bytes.Buffer, err error) {
			// block that triggers if the do.Path was NOT set
			//   via cli flag... or not
			fname := filepath.Join(args.Path, work.playbook)

			// if YAMLPath cannot be found, abort
			if _, err := os.Stat(fname); os.IsNotExist(err) {
				return logBuf, fmt.Errorf("could not find playbook file (%s)",
					fname)
			}

			if args.Jobs != 1 {
				logger = logging.NewLogger(log.NewLogfmtLogger(&logBuf))
				if !args.Debug {
					logger.Trace = log.NewNopLogger()
				}
			}

			// Load the package if it doesn't exist
			script, err := loader.LoadPlaybook(fname, args, logger)
			if err != nil {
				return logBuf, err
			}

			// Load existing bin files to decode events
			var abiError error
			client.AllSpecs, abiError = abi.LoadPath(script.BinPath)
			if err != nil {
				logger.InfoMsg("failed to load ABIs for Event parsing", "path", script.BinPath, "error", abiError)
			}

			err = jobs.ExecutePlaybook(args, script, client, logger)
			return
		}

		startTime := time.Now()
		logBuf, err := doWork(playbook)
		results <- playbookResult{
			jobNo:    playbook.jobNo,
			log:      logBuf,
			err:      err,
			duration: time.Since(startTime),
		}
	}
}

// RunPlaybooks starts workers, and loads the playbooks in parallel in the workers, and executes them.
func RunPlaybooks(args *def.DeployArgs, playbooks []string, logger *logging.Logger) (int, error) {
	// if bin and abi paths are default cli settings then use the
	//   stated defaults of do.Path plus bin|abi
	if args.Path == "" {
		var err error
		args.Path, err = os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("failed to get current directory %v", err))
		}
	}

	// useful for debugging
	logger.InfoMsg("Using chain", "Chain", args.Chain, "Signer", args.KeysService)

	workQ := make(chan playbookWork, 100)
	resultQ := make(chan playbookResult, 100)

	for i := 1; i <= args.Jobs; i++ {
		go worker(workQ, resultQ, args, logger)
	}

	for i, playbook := range playbooks {
		workQ <- playbookWork{jobNo: i, playbook: playbook}
	}
	close(workQ)

	// work queued, now read the results
	results := make([]*playbookResult, len(playbooks))
	printed := 0
	failures := 0
	successes := 0

	for range playbooks {
		// Receive results as they come
		jobResult := <-resultQ
		results[jobResult.jobNo] = &jobResult
		// Print them in order
		for results[printed] != nil {
			res := results[printed]
			os.Stderr.Write(res.log.Bytes())
			if res.err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v", res.err)
			}
			res.log.Truncate(0)
			if res.err != nil {
				failures++
			} else {
				successes++
			}
			printed++
			if printed >= len(playbooks) {
				break
			}
		}
	}
	close(resultQ)

	if successes > 0 {
		logger.InfoMsg("JOBS THAT SUCCEEDED", "count", successes)
		for i, playbook := range playbooks {
			res := results[i]
			if res.err != nil {
				continue
			}
			logger.InfoMsg("Playbook result",
				"jobNo", i,
				"file", playbook,
				"time", res.duration.String())
		}
	}

	if failures > 0 {
		logger.InfoMsg("JOBS THAT FAILED", "count", failures)
		for i, playbook := range playbooks {
			res := results[i]
			if res.err == nil {
				continue
			}
			logger.InfoMsg("Playbook result",
				"jobNo", i,
				"file", playbook,
				"error", res.err,
				"time", res.duration.String())
		}
	}

	return failures, nil
}
