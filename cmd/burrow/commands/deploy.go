package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	pkgs "github.com/hyperledger/burrow/deploy"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/proposals"
	"github.com/hyperledger/burrow/logging"
	cli "github.com/jawher/mow.cli"
)

// 15 seconds is like a long time man
const defaultChainTimeout = 15 * time.Second

// Deploy runs the desired playbook(s)
func Deploy(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainOpt := cmd.StringOpt("c chain", "127.0.0.1:10997", "chain to be used in IP:PORT format")

		signerOpt := cmd.StringOpt("s keys", "",
			"IP:PORT of Burrow GRPC service which jobs should or otherwise transaction submitted unsigned for mempool signing in Burrow")

		mempoolSigningOpt := cmd.BoolOpt("p mempool-signing", false,
			"Use Burrow's own keys connection to sign transactions - means that Burrow instance must have access to input account keys. "+
				"Sequence numbers are set as transactions enter the mempool so concurrent transactions can be sent from same inputs.")

		pathOpt := cmd.StringOpt("i dir", "", "root directory of app (will use pwd by default)")

		defaultOutputOpt := cmd.StringOpt("o output", def.DefaultOutputFile,
			"filename for playbook output file. by default, this name will reflect the playbook passed")

		playbooksOpt := cmd.StringsArg("FILE", []string{},
			"path to playbook file which deploy should run. if also using the --dir flag, give the relative path to playbooks file, which should be in the same directory")

		defaultSetsOpt := cmd.StringsOpt("e set", []string{},
			"default sets to use; operates the same way as the [set] jobs, only before the jobs file is ran (and after default address")

		binPathOpt := cmd.StringOpt("b bin-path", "[dir]/bin",
			"path to the bin directory jobs should use when saving binaries after the compile process defaults to --dir + /bin")

		defaultGasOpt := cmd.StringOpt("g gas", "1111111111",
			"default gas to use; can be overridden for any single job")

		jobsOpt := cmd.IntOpt("j jobs", 1,
			"default number of concurrent playbooks to run if multiple are specified")

		addressOpt := cmd.StringOpt("a address", "",
			"default address (or account name) to use; operates the same way as the [account] job, only before the deploy file is ran")

		defaultFeeOpt := cmd.StringOpt("n fee", "9999", "default fee to use")

		defaultAmountOpt := cmd.StringOpt("m amount", "9999",
			"default amount to use")

		verboseOpt := cmd.BoolOpt("v verbose", false, "verbose output")

		localAbiOpt := cmd.BoolOpt("local-abi", false, "use local ABIs rather than fetching them from burrow")

		wasmOpt := cmd.BoolOpt("wasm", false, "Compile to WASM using solang (experimental)")

		debugOpt := cmd.BoolOpt("d debug", false, "debug level output")

		proposalVerify := cmd.BoolOpt("proposal-verify", false, "Verify any proposal, do NOT create new proposal or vote")

		proposalVote := cmd.BoolOpt("proposal-vote", false, "Vote for proposal, do NOT create new proposal")

		proposalCreate := cmd.BoolOpt("proposal-create", false, "Create new proposal")

		timeoutSecondsOpt := cmd.IntOpt("t timeout", int(defaultChainTimeout/time.Second), "Timeout to talk to the chain in seconds")

		proposalList := cmd.StringOpt("list-proposals state", "", "List proposals, either all, executed, expired, or current")

		cmd.Spec = "[--chain=<host:port>] [--keys=<host:port>] [--mempool-signing] [--dir=<root directory>] " +
			"[--output=<output file>] [--wasm] [--set=<KEY=VALUE>]... [--bin-path=<path>] [--gas=<gas>] " +
			"[--jobs=<concurrent playbooks>] [--address=<address>] [--fee=<fee>] [--amount=<amount>] [--local-abi] " +
			"[--verbose] [--debug] [--timeout=<timeout>] " +
			"[--list-proposals=<state> | --proposal-create| --proposal-verify | --proposal-vote] [FILE...]"

		cmd.Action = func() {
			args := new(def.DeployArgs)

			if *proposalVerify && *proposalVote {
				output.Fatalf("Cannot combine --proposal-verify and --proposal-vote")
			}

			for _, e := range *defaultSetsOpt {
				s := strings.Split(e, "=")
				if len(s) != 2 || s[0] == "" {
					output.Fatalf("`--set %s' should have format VARIABLE=value", e)
				}
			}

			args.Chain = *chainOpt
			args.KeysService = *signerOpt
			args.MempoolSign = *mempoolSigningOpt
			args.Timeout = *timeoutSecondsOpt
			args.Path = *pathOpt
			args.LocalABI = *localAbiOpt
			args.Wasm = *wasmOpt
			args.DefaultOutput = *defaultOutputOpt
			args.DefaultSets = *defaultSetsOpt
			args.BinPath = *binPathOpt
			args.DefaultGas = *defaultGasOpt
			args.Address = *addressOpt
			args.DefaultFee = *defaultFeeOpt
			args.DefaultAmount = *defaultAmountOpt
			args.Verbose = *verboseOpt
			args.Debug = *debugOpt
			args.Jobs = *jobsOpt
			args.ProposeVerify = *proposalVerify
			args.ProposeVote = *proposalVote
			args.ProposeCreate = *proposalCreate
			stderrLogger := log.NewLogfmtLogger(os.Stderr)
			logger := logging.NewLogger(stderrLogger)
			handleTerm()

			if !*debugOpt {
				logger.Trace = log.NewNopLogger()
			}

			if *proposalList != "" {
				state, err := proposals.ProposalStateFromString(*proposalList)
				if err != nil {
					output.Fatalf(err.Error())
				}
				err = proposals.ListProposals(args, state, logger)
				if err != nil {
					output.Fatalf(err.Error())
				}
			} else {
				if len(*playbooksOpt) == 0 {
					output.Fatalf("incorrect usage: missing deployment yaml file(s)")
				}
				failures, err := pkgs.RunPlaybooks(args, *playbooksOpt, logger)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				if failures > 0 {
					os.Exit(failures)
				}
			}
		}
	}
}
