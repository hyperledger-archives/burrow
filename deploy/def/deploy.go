package def

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/deploy/def/rule"
)

const DefaultOutputFile = "deploy.output.json"

type DeployArgs struct {
	Chain         string   `mapstructure:"," json:"," yaml:"," toml:","`
	KeysService   string   `mapstructure:"," json:"," yaml:"," toml:","`
	MempoolSign   bool     `mapstructure:"," json:"," yaml:"," toml:","`
	LocalABI      bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Wasm          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Timeout       int      `mapstructure:"," json:"," yaml:"," toml:","`
	Address       string   `mapstructure:"," json:"," yaml:"," toml:","`
	BinPath       string   `mapstructure:"," json:"," yaml:"," toml:","`
	CurrentOutput string   `mapstructure:"," json:"," yaml:"," toml:","`
	Debug         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAmount string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultFee    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultGas    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultOutput string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultSets   []string `mapstructure:"," json:"," yaml:"," toml:","`
	Path          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Verbose       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Jobs          int      `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeVerify bool     `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeVote   bool     `mapstructure:"," json:"," yaml:"," toml:","`
	ProposeCreate bool     `mapstructure:"," json:"," yaml:"," toml:","`
}

func (args *DeployArgs) Validate() error {
	return validation.ValidateStruct(args,
		validation.Field(&args.DefaultAmount, rule.Uint64),
		validation.Field(&args.DefaultFee, rule.Uint64),
		validation.Field(&args.DefaultGas, rule.Uint64),
	)
}

type Playbook struct {
	Filename string
	Account  string
	Jobs     []*Job
	Path     string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	BinPath  string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	// If we're in a proposal or meta job, reference our parent script
	Parent *Playbook `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
}

func (pkg *Playbook) Validate() error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Jobs),
	)
}
