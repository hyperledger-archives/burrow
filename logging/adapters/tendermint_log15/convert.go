package adapters

import (
	"time"

	. "github.com/eris-ltd/eris-db/util/slice"
	"github.com/go-stack/stack"
	"github.com/tendermint/log15"
	"github.com/eris-ltd/eris-db/logging/structure"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"fmt"
)

// Convert a go-kit log line (i.e. keyvals... interface{}) into a log15 record
// This allows us to use log15 output handlers
func LogLineToRecord(keyvals... interface{}) *log15.Record {
	vals, ctx := structure.ValuesAndContext(keyvals, structure.TimeKey,
		structure.MessageKey, structure.CallerKey, structure.LevelKey)

	theTime, _ := vals[structure.TimeKey].(time.Time)
	call, _ := vals[structure.CallerKey].(stack.Call)
	level, _ := vals[structure.LevelKey].(string)
	message, _ := vals[structure.MessageKey].(string)

	fmt.Println(keyvals...)
	return &log15.Record{
		Time: theTime,
		Lvl:  Log15LvlFromString(level),
		Msg:  message,
		Call: call,
		Ctx:  append(ctx, structure.CallerKey, call),
		KeyNames: log15.RecordKeyNames{
			Time: structure.TimeKey,
			Msg:  structure.MessageKey,
			Lvl:  structure.LevelKey,
		}}
}

// Convert a log15 record to a go-kit log line (i.e. keyvals... interface{})
// This allows us to capture output from dependencies using log15
func RecordToLogLine(record *log15.Record) []interface{} {
	return Concat(
		Slice(
			structure.TimeKey, record.Time,
			structure.CallerKey, record.Call,
			structure.LevelKey, record.Lvl.String(),
		),
		record.Ctx,
		Slice(
			structure.MessageKey, record.Msg,
		))
}

// Collapse our weak notion of leveling and log15's into a log15.Lvl
func Log15LvlFromString(level string) log15.Lvl {
	if level == "" {
		return log15.LvlDebug
	}
	switch level {
	case loggers.InfoLevelName:
		return log15.LvlInfo
	case loggers.TraceLevelName:
		return log15.LvlDebug
	default:
		lvl, err := log15.LvlFromString(level)
		if err == nil {
			return lvl
		}
		return log15.LvlDebug
	}
}