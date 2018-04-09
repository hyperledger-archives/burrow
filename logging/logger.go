// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	kitlog "github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/structure"
)

// InfoTraceLogger maintains provides two logging 'channels' that are interlaced
// to provide a coarse grained filter to distinguish human-consumable 'Info'
// messages and execution level 'Trace' messages.
type Logger struct {
	// Send a log message to the Info channel, formed of a sequence of key value
	// pairs. Info messages should be operationally interesting to a human who is
	// monitoring the logs. But not necessarily a human who is trying to
	// understand or debug the system. Any handled errors or warnings should be
	// sent to the Info channel (where you may wish to tag them with a suitable
	// key-value pair to categorise them as such).
	Info kitlog.Logger
	// Send an log message to the Trace channel, formed of a sequence of key-value
	// pairs. Trace messages can be used for any state change in the system that
	// may be of interest to a machine consumer or a human who is trying to debug
	// the system or trying to understand the system in detail. If the messages
	// are very point-like and contain little structure, consider using a metric
	// instead.
	Trace  kitlog.Logger
	Output *kitlog.SwapLogger
}

// Create an InfoTraceLogger by passing the initial outputLogger.
func NewLogger(outputLogger kitlog.Logger) *Logger {
	// We will never halt the progress of a log emitter. If log output takes too
	// long will start dropping log lines by using a ring buffer.
	swapLogger := new(kitlog.SwapLogger)
	swapLogger.Swap(outputLogger)
	return &Logger{
		Output: swapLogger,
		// logging contexts
		Info: kitlog.With(swapLogger,
			structure.ChannelKey, structure.InfoChannelName,
		),
		Trace: kitlog.With(swapLogger,
			structure.ChannelKey, structure.TraceChannelName,
		),
	}
}

func NewNoopLogger() *Logger {
	return &Logger{
		Info:   kitlog.NewNopLogger(),
		Trace:  kitlog.NewNopLogger(),
		Output: new(kitlog.SwapLogger),
	}
}

// A logging context (see go-kit log's Context). Takes a sequence key values
// via With or WithPrefix and ensures future calls to log will have those
// contextual values appended to the call to an underlying logger.
// Values can be dynamic by passing an instance of the kitlog.Valuer interface
// This provides an interface version of the kitlog.Context struct to be used
// For implementations that wrap a kitlog.Context. In addition it makes no
// assumption about the name or signature of the logging method(s).
// See InfoTraceLogger
func (l *Logger) With(keyvals ...interface{}) *Logger {
	return &Logger{
		Output: l.Output,
		Info:   kitlog.With(l.Info, keyvals...),
		Trace:  kitlog.With(l.Trace, keyvals...),
	}
}

// Establish a context on the Info channel keeping Trace the same
func (l *Logger) WithInfo(keyvals ...interface{}) *Logger {
	return &Logger{
		Output: l.Output,
		Info:   kitlog.With(l.Info, keyvals...),
		Trace:  l.Trace,
	}
}

// Establish a context on the Trace channel keeping Info the same
func (l *Logger) WithTrace(keyvals ...interface{}) *Logger {
	return &Logger{
		Output: l.Output,
		Info:   l.Info,
		Trace:  kitlog.With(l.Trace, keyvals...),
	}
}

func (l *Logger) WithPrefix(keyvals ...interface{}) *Logger {
	return &Logger{
		Output: l.Output,
		Info:   kitlog.WithPrefix(l.Info, keyvals...),
		Trace:  kitlog.WithPrefix(l.Trace, keyvals...),
	}
}

// Hot swap the underlying outputLogger with another one to re-route messages
func (l *Logger) SwapOutput(infoLogger kitlog.Logger) {
	l.Output.Swap(infoLogger)
}

// Record structured Info lo`g line with a message
func (l *Logger) InfoMsg(message string, keyvals ...interface{}) error {
	return Msg(l.Info, message, keyvals...)
}

// Record structured Trace log line with a message
func (l *Logger) TraceMsg(message string, keyvals ...interface{}) error {
	return Msg(l.Trace, message, keyvals...)
}

// Establish or extend the scope of this logger by appending scopeName to the Scope vector.
// Like With the logging scope is append only but can be used to provide parenthetical scopes by hanging on to the
// parent scope and using once the scope has been exited. The scope mechanism does is agnostic to the type of scope
// so can be used to identify certain segments of the call stack, a lexical scope, or any other nested scope.
func (l *Logger) WithScope(scopeName string) *Logger {
	// InfoTraceLogger will collapse successive (ScopeKey, scopeName) pairs into a vector in the order which they appear
	return l.With(structure.ScopeKey, scopeName)
}

// Record a structured log line with a message
func Msg(logger kitlog.Logger, message string, keyvals ...interface{}) error {
	prepended := structure.CopyPrepend(keyvals, structure.MessageKey, message)
	return logger.Log(prepended...)
}
