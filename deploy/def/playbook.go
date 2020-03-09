package def

import "github.com/go-ozzo/ozzo-validation"

type Playbook struct {
	Filename string
	Account  string
	// Prevent this playbook from running at the same time as other playbooks
	NoParallel bool `mapstructure:"no-parallel,omitempty" json:"no-parallel,omitempty" yaml:"no-parallel,omitempty" toml:"no-parallel,omitempty"`
	Jobs       []*Job
	Path       string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	BinPath    string `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	// If we're in a proposal or meta job, reference our parent script
	Parent *Playbook `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
}

func (pkg *Playbook) Validate() error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Jobs),
	)
}
