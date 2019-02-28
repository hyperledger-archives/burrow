package commands

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	pkgs "github.com/hyperledger/burrow/deploy"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/proposals"
	"github.com/hyperledger/burrow/deploy/util"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// 15 seconds is like a long time man
const defaultChainTimeout = 15 * time.Second

func Deploy(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainOpt := cmd.StringOpt("u chain", "127.0.0.1:10997", "chain to be used in IP:PORT format")

		signerOpt := cmd.StringOpt("s keys", "",
			"IP:PORT of Burrow GRPC service which jobs should or otherwise transaction submitted unsigned for mempool signing in Burrow")

		mempoolSigningOpt := cmd.BoolOpt("p mempool-signing", false,
			"Use Burrow's own keys connection to sign transactions - means that Burrow instance must have access to input account keys. "+
				"Sequence numbers are set as transactions enter the mempool so concurrent transactions can be sent from same inputs.")

		pathOpt := cmd.StringOpt("i dir", "", "root directory of app (will use pwd by default)")

		defaultOutputOpt := cmd.StringOpt("o output", def.DefaultOutputFile,
			"filename for playbook output file. by default, this name will reflect the playbook passed")

		playbooksOpt := cmd.StringsArg("FILE", []string{"deploy.yaml"},
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
			"default address to use; operates the same way as the [account] job, only before the deploy file is ran")

		defaultFeeOpt := cmd.StringOpt("n fee", "9999", "default fee to use")

		defaultAmountOpt := cmd.StringOpt("m amount", "9999",
			"default amount to use")

		verboseOpt := cmd.BoolOpt("v verbose", false, "verbose output")

		debugOpt := cmd.BoolOpt("d debug", false, "debug level output")

		proposalVerify := cmd.BoolOpt("proposal-verify", false, "Verify any proposal, do NOT create new proposal or vote")

		proposalVote := cmd.BoolOpt("proposal-vote", false, "Vote for proposal, do NOT create new proposal")

		proposalCreate := cmd.BoolOpt("proposal-create", false, "Create new proposal")

		timeoutSecondsOpt := cmd.IntOpt("t timeout", int(defaultChainTimeout/time.Second), "Timeout to talk to the chain in seconds")

		proposalList := cmd.StringOpt("list-proposals state", "", "List proposals, either all, executed, expired, or current")

		cmd.Spec = "[--chain=<host:port>] [--keys=<host:port>] [--mempool-signing] [--dir=<root directory>] " +
			"[--output=<output file>] [--set=<KEY=VALUE>]... [--bin-path=<path>] [--gas=<gas>] " +
			"[--jobs=<concurrent playbooks>] [--address=<address>] [--fee=<fee>] [--amount=<amount>] " +
			"[--verbose] [--debug] [--timeout=<timeout>] [--proposal-create|--proposal-verify|--proposal-create] [FILE...]"

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
			log.SetFormatter(new(PlainFormatter))
			log.SetLevel(log.WarnLevel)
			if args.Verbose {
				log.SetLevel(log.InfoLevel)
			} else if args.Debug {
				log.SetLevel(log.DebugLevel)
			}
			handleTerm()

			if *proposalList != "" {
				state, err := proposals.ProposalStateFromString(*proposalList)
				if err != nil {
					output.Fatalf(err.Error())
				}
				proposals.ListProposals(args, state)
			} else {
				util.IfExit(pkgs.RunPlaybook(args, *playbooksOpt))
			}
		}
	}
}

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.appendMessage(b, entry.Message)
	for _, key := range keys {
		f.appendMessageData(b, key, entry.Data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *PlainFormatter) appendMessage(b *bytes.Buffer, message string) {
	fmt.Fprintf(b, "%-44s", message)
}

func (f *PlainFormatter) appendMessageData(b *bytes.Buffer, key string, value interface{}) {
	switch key {
	case "":
		b.WriteString("=> ")
	case "=>":
		b.WriteString(key)
		b.WriteByte(' ')
	default:
		b.WriteString(key)
		b.WriteString(" => ")
	}
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(stringVal)
	b.WriteString(" ")
}
