package def

import (
	"github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/deploy/def/rule"
)

type Packages struct {
	Address       string   `mapstructure:"," json:"," yaml:"," toml:","`
	BinPath       string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainURL      string   `mapstructure:"," json:"," yaml:"," toml:","`
	CurrentOutput string   `mapstructure:"," json:"," yaml:"," toml:","`
	Debug         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAmount string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultFee    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultGas    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultOutput string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultSets   []string `mapstructure:"," json:"," yaml:"," toml:","`
	Path          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Signer        string   `mapstructure:"," json:"," yaml:"," toml:","`
	Verbose       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	YAMLPath      string   `mapstructure:"," json:"," yaml:"," toml:","`
	Jobs          int      `mapstructure:"," json:"," yaml:"," toml:","`

	Package *Package
	Client
}

func (do *Packages) Validate() error {
	return validation.ValidateStruct(do,
		validation.Field(&do.Address, rule.Address),
		validation.Field(&do.DefaultAmount, rule.Uint64),
		validation.Field(&do.DefaultFee, rule.Uint64),
		validation.Field(&do.DefaultGas, rule.Uint64),
		validation.Field(&do.Package),
	)
}

func (do *Packages) Dial() error {
	return do.Client.Dial(do.ChainURL, do.Signer)
}

type Package struct {
	Account string
	Jobs    []*Job
}

func (pkg *Package) Validate() error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Account, rule.Address),
		validation.Field(&pkg.Jobs),
	)
}
