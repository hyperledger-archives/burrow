package commands

import (
	"fmt"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis/spec"
	cli "github.com/jawher/mow.cli"
)

func Spec(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		tomlOpt := cmd.BoolOpt("t toml", false, "Emit GenesisSpec as TOML rather than the "+
			"default JSON")

		baseSpecsArg := cmd.StringsArg("BASE", nil, "Provide a base GenesisSpecs on top of which any "+
			"additional GenesisSpec presets specified by other flags will be merged. GenesisSpecs appearing "+
			"later take precedent over those appearing early if multiple --base flags are provided")

		accountNamePrefixOpt := cmd.StringOpt("x name-prefix", "", "Prefix added to the names of accounts in GenesisSpec")
		fullOpt := cmd.IntOpt("f full-accounts", 0, "Number of preset Full type accounts")
		validatorOpt := cmd.IntOpt("v validator-accounts", 0, "Number of preset Validator type accounts")
		rootOpt := cmd.IntOpt("r root-accounts", 0, "Number of preset Root type accounts")
		developerOpt := cmd.IntOpt("d developer-accounts", 0, "Number of preset Developer type accounts")
		participantsOpt := cmd.IntOpt("p participant-accounts", 0, "Number of preset Participant type accounts")
		chainNameOpt := cmd.StringOpt("n chain-name", "", "Default chain name")

		cmd.Spec = "[--name-prefix=<prefix for account names>][--full-accounts] [--validator-accounts] [--root-accounts] " +
			"[--developer-accounts] [--participant-accounts] [--chain-name] [--toml] [BASE...]"

		cmd.Action = func() {
			specs := make([]spec.GenesisSpec, 0, *participantsOpt+*fullOpt)
			for _, baseSpec := range *baseSpecsArg {
				genesisSpec := new(spec.GenesisSpec)
				err := source.FromFile(baseSpec, genesisSpec)
				if err != nil {
					output.Fatalf("could not read GenesisSpec: %v", err)
				}
				specs = append(specs, *genesisSpec)
			}
			for i := 0; i < *fullOpt; i++ {
				specs = append(specs, spec.FullAccount(fmt.Sprintf("%sFull_%v", *accountNamePrefixOpt, i)))
			}
			for i := 0; i < *validatorOpt; i++ {
				specs = append(specs, spec.ValidatorAccount(fmt.Sprintf("%sValidator_%v", *accountNamePrefixOpt, i)))
			}
			for i := 0; i < *rootOpt; i++ {
				specs = append(specs, spec.RootAccount(fmt.Sprintf("%sRoot_%v", *accountNamePrefixOpt, i)))
			}
			for i := 0; i < *developerOpt; i++ {
				specs = append(specs, spec.DeveloperAccount(fmt.Sprintf("%sDeveloper_%v", *accountNamePrefixOpt, i)))
			}
			for i := 0; i < *participantsOpt; i++ {
				specs = append(specs, spec.ParticipantAccount(fmt.Sprintf("%sParticipant_%v", *accountNamePrefixOpt, i)))
			}
			genesisSpec := spec.MergeGenesisSpecs(specs...)
			if *chainNameOpt != "" {
				genesisSpec.ChainName = *chainNameOpt
			}
			if *tomlOpt {
				output.Printf(source.TOMLString(genesisSpec))
			} else {
				output.Printf(source.JSONString(genesisSpec))
			}
		}
	}
}
