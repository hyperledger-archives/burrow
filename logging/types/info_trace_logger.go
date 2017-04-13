package types

import kitlog "github.com/go-kit/kit/log"

const (
	InfoChannelName  = "Info"
	TraceChannelName = "Trace"

	InfoLevelName  = InfoChannelName
	TraceLevelName = TraceChannelName
)

// InfoTraceLogger maintains two independent concurrently-safe channels of
// logging. The idea behind the independence is that you can ignore one channel
// with no performance penalty. For more fine grained filtering or aggregation
// the Info and Trace loggers can be decorated loggers that perform arbitrary
// filtering/routing/aggregation on log messages.
type InfoTraceLogger interface {
	// Send a log message to the default channel
	kitlog.Logger

	// Send an log message to the Info channel, formed of a sequence of key value
	// pairs. Info messages should be operationally interesting to a human who is
	// monitoring the logs. But not necessarily a human who is trying to
	// understand or debug the system. Any handled errors or warnings should be
	// sent to the Info channel (where you may wish to tag them with a suitable
	// key-value pair to categorise them as such). Info messages will be sent to
	// the infoOnly and the infoAndTrace output loggers.
	Info(keyvals ...interface{}) error

	// Send an log message to the Trace channel, formed of a sequence of key-value
	// pairs. Trace messages can be used for any state change in the system that
	// may be of interest to a machine consumer or a human who is trying to debug
	// the system or trying to understand the system in detail. If the messages
	// are very point-like and contain little structure, consider using a metric
	// instead. Trace messages will only be sent to the infoAndTrace output logger.
	Trace(keyvals ...interface{}) error

	// A logging context (see go-kit log's Context). Takes a sequence key values
	// via With or WithPrefix and ensures future calls to log will have those
	// contextual values appended to the call to an underlying logger.
	// Values can be dynamic by passing an instance of the kitlog.Valuer interface
	// This provides an interface version of the kitlog.Context struct to be used
	// For implementations that wrap a kitlog.Context. In addition it makes no
	// assumption about the name or signature of the logging method(s).
	// See InfoTraceLogger

	// Establish a context by appending contextual key-values to any existing
	// contextual values
	With(keyvals ...interface{}) InfoTraceLogger

	// Establish a context by prepending contextual key-values to any existing
	// contextual values
	WithPrefix(keyvals ...interface{}) InfoTraceLogger

	// Hot swap the underlying infoOnlyLogger with another one to re-route messages
	// on the Info channel
	SwapInfoOnlyOutput(infoOnlyLogger kitlog.Logger)

	// Hot swap the underlying infoAndTraceLogger with another one to re-route messages
	// on the Info and Trace channels
	SwapInfoAndTraceOutput(infoTraceLogger kitlog.Logger)
}

// Interface assertions
var _ kitlog.Logger = (InfoTraceLogger)(nil)
