package presets

import (
	"fmt"

	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"
)

// Function to generate part of a tree of Sinks (e.g. append a single child node, or an entire subtree).
// Each Instruction takes a (sub)root, which it may modify, and appends further child sinks below that root.
// Returning the new subroot from which to apply any further Presets.
// When chained together in a pre-order instructions can be composed to form an entire Sink tree.
type Instruction struct {
	name  string
	desc  string
	nargs int
	// The builder for the Instruction is a function that may modify the stack or ops string. Typically
	// by mutating the sink at the top of the stack and may move the cursor or by pushing child sinks
	// to the stack. The builder may also return a modified ops slice whereby it may insert Instruction calls
	// acting as a macro or consume ops as arguments.
	builder func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error)
}

func (i Instruction) Name() string {
	return i.name
}

func (i Instruction) Description() string {
	return i.desc
}

const (
	Info       = "info"
	Minimal    = "minimal"
	IncludeAny = "include-any"
	Stderr     = "stderr"
	Stdout     = "stdout"
	Terminal   = "terminal"
	JSON       = "json"
	Up         = "up"
	Down       = "down"
	File       = "file"
)

var instructions = []Instruction{
	{
		name: Up,
		desc: "Ascend the sink tree by travelling up the stack to the previous sink recorded on the stack",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			return pop(stack), nil
		},
	},
	{
		name: Down,
		desc: "Descend the sink tree by inserting a sink as a child to the current sink and adding it to the stack",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			return push(stack, config.Sink()), nil
		},
	},
	{
		name: Minimal,
		desc: "A generally less chatty log output, follow with output options",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			return push(stack,
				config.Sink().SetTransform(config.PruneTransform(structure.TraceKey, structure.RunId)),
				config.Sink().SetTransform(config.FilterTransform(config.IncludeWhenAllMatch,
					structure.ChannelKey, structure.InfoChannelName)),
				config.Sink().SetTransform(config.FilterTransform(config.ExcludeWhenAnyMatches,
					structure.ComponentKey, "Tendermint",
					"module", "p2p",
					"module", "mempool"))), nil
		},
	},
	{
		name: IncludeAny,
		desc: "Establish an 'include when any predicate matches' filter transform at this this sink",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureFilter(sink)
			sink.Transform.FilterConfig.FilterMode = config.IncludeWhenAnyMatches
			return stack, nil
		},
	},
	{
		name: Info,
		desc: "Add a filter predicate to match the Info logging channel",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureFilter(sink)
			sink.Transform.FilterConfig.AddPredicate(structure.ChannelKey, structure.InfoChannelName)
			return stack, nil
		},
	},
	{
		name: Stdout,
		desc: "Use Stdout output for this sink",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureOutput(sink)
			sink.Output.OutputType = config.Stdout
			return stack, nil
		},
	},
	{
		name: Stderr,
		desc: "Use Stderr output for this sink",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureOutput(sink)
			sink.Output.OutputType = config.Stderr
			return stack, nil
		},
	},
	{
		name: Terminal,
		desc: "Use the the terminal output format for this sink",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureOutput(sink)
			sink.Output.Format = loggers.TerminalFormat
			return stack, nil
		},
	},
	{
		name: JSON,
		desc: "Use the the terminal output format for this sink",
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureOutput(sink)
			sink.Output.Format = loggers.JSONFormat
			return stack, nil
		},
	},
	{
		name:  File,
		desc:  "Use the the terminal output format for this sink",
		nargs: 1,
		builder: func(stack []*config.SinkConfig, args []string) ([]*config.SinkConfig, error) {
			sink := peek(stack)
			ensureOutput(sink)
			sink.Output.OutputType = config.File
			sink.Output.FileConfig = &config.FileConfig{
				Path: args[0],
			}
			return stack, nil
		},
	},
}

var instructionsMap map[string]Instruction

func init() {
	instructionsMap = make(map[string]Instruction, len(instructions))
	for _, p := range instructions {
		instructionsMap[p.name] = p
	}
}

func Instructons() []Instruction {
	ins := make([]Instruction, len(instructions))
	copy(ins, instructions)
	return ins
}

func Describe(name string) string {
	preset, ok := instructionsMap[name]
	if !ok {
		return fmt.Sprintf("No logging preset named '%s'", name)
	}
	return preset.desc
}

func BuildSinkConfig(ops ...string) (*config.SinkConfig, error) {
	stack := []*config.SinkConfig{config.Sink()}
	var err error
	pos := 0
	for len(ops) > 0 {
		// Keep applying instructions until their are no ops left
		instruction, ok := instructionsMap[ops[0]]
		if !ok {
			return nil, fmt.Errorf("could not find logging preset '%s'", ops[0])
		}
		// pop instruction name
		ops = ops[1:]
		if len(ops) < instruction.nargs {
			return nil, fmt.Errorf("did not have enough arguments for instruction %s at position %v "+
				"(requires %v arguments)", instruction.name, pos, instruction.nargs)
		}
		stack, err = instruction.builder(stack, ops[:instruction.nargs])
		if err != nil {
			return nil, err
		}
		// pop instruction args
		ops = ops[instruction.nargs:]
		pos++
	}
	return stack[0], nil
}

func ensureFilter(sinkConfig *config.SinkConfig) {
	if sinkConfig.Transform == nil {
		sinkConfig.Transform = &config.TransformConfig{}
	}
	if sinkConfig.Transform.FilterConfig == nil {
		sinkConfig.Transform.FilterConfig = &config.FilterConfig{}
	}
	sinkConfig.Transform.TransformType = config.Filter
}

func ensureOutput(sinkConfig *config.SinkConfig) {
	if sinkConfig.Output == nil {
		sinkConfig.Output = &config.OutputConfig{}
	}
}

// Push a path sequence of sinks onto the stack
func push(stack []*config.SinkConfig, sinkConfigs ...*config.SinkConfig) []*config.SinkConfig {
	for _, sinkConfig := range sinkConfigs {
		peek(stack).AddSinks(sinkConfig)
		stack = append(stack, sinkConfig)
	}
	return stack
}

func pop(stack []*config.SinkConfig) []*config.SinkConfig {
	return stack[:len(stack)-1]
}

func peek(stack []*config.SinkConfig) *config.SinkConfig {
	return stack[len(stack)-1]
}
