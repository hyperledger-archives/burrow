package def

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/deploy/def/rule"
	"github.com/hyperledger/burrow/execution/evm/abi"
)

const DefaultOutputFile = "deploy.output.json"

type DeployArgs struct {
	Address       string       `mapstructure:"," json:"," yaml:"," toml:","`
	BinPath       string       `mapstructure:"," json:"," yaml:"," toml:","`
	CurrentOutput string       `mapstructure:"," json:"," yaml:"," toml:","`
	Debug         bool         `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAmount string       `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultFee    string       `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultGas    string       `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultOutput string       `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultSets   []string     `mapstructure:"," json:"," yaml:"," toml:","`
	Path          string       `mapstructure:"," json:"," yaml:"," toml:","`
	Verbose       bool         `mapstructure:"," json:"," yaml:"," toml:","`
	YAMLPath      string       `mapstructure:"," json:"," yaml:"," toml:","`
	Jobs          int          `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeVerify bool         `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeVote   bool         `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeCreate bool         `mapstructure:"," json:"," yaml:"," toml:","`
	AllSpecs      *abi.AbiSpec `mapstructure:"," json:"," yaml:"," toml:","`
}

func (do *DeployArgs) Validate() error {
	return validation.ValidateStruct(do,
		validation.Field(&do.DefaultAmount, rule.Uint64),
		validation.Field(&do.DefaultFee, rule.Uint64),
		validation.Field(&do.DefaultGas, rule.Uint64),
	)
}

type Playbook struct {
	Account string
	Jobs    []*Job
	Path    string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	BinPath string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	// If we're in a proposal or meta job, reference our parent script
	Parent *Playbook `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
}

func (pkg *Playbook) Validate() error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Jobs),
	)
}
