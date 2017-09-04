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

package tendermint_log15

import (
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-stack/stack"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/logging/types"
	. "github.com/hyperledger/burrow/util/slice"
	"github.com/tendermint/log15"
)

type infoTraceLoggerAsLog15Handler struct {
	logger types.InfoTraceLogger
}

var _ log15.Handler = (*infoTraceLoggerAsLog15Handler)(nil)

type log15HandlerAsKitLogger struct {
	handler log15.Handler
}

var _ kitlog.Logger = (*log15HandlerAsKitLogger)(nil)

func (l *log15HandlerAsKitLogger) Log(keyvals ...interface{}) error {
	record := LogLineToRecord(keyvals...)
	return l.handler.Log(record)
}

func (h *infoTraceLoggerAsLog15Handler) Log(record *log15.Record) error {
	if record.Lvl < log15.LvlDebug {
		// Send to Critical, Warning, Error, and Info to the Info channel
		h.logger.Info(RecordToLogLine(record)...)
	} else {
		// Send to Debug to the Trace channel
		h.logger.Trace(RecordToLogLine(record)...)
	}
	return nil
}

func Log15HandlerAsKitLogger(handler log15.Handler) kitlog.Logger {
	return &log15HandlerAsKitLogger{
		handler: handler,
	}
}

func InfoTraceLoggerAsLog15Handler(logger types.InfoTraceLogger) log15.Handler {
	return &infoTraceLoggerAsLog15Handler{
		logger: logger,
	}
}

// Convert a go-kit log line (i.e. keyvals... interface{}) into a log15 record
// This allows us to use log15 output handlers
func LogLineToRecord(keyvals ...interface{}) *log15.Record {
	vals, ctx := structure.ValuesAndContext(keyvals, structure.TimeKey,
		structure.MessageKey, structure.CallerKey, structure.LevelKey)

	// Mapping of log line to Record is on a best effort basis
	theTime, _ := vals[structure.TimeKey].(time.Time)
	call, _ := vals[structure.CallerKey].(stack.Call)
	level, _ := vals[structure.LevelKey].(string)
	message, _ := vals[structure.MessageKey].(string)

	return &log15.Record{
		Time: theTime,
		Lvl:  Log15LvlFromString(level),
		Msg:  message,
		Call: call,
		Ctx:  ctx,
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
	case types.InfoLevelName:
		return log15.LvlInfo
	case types.TraceLevelName:
		return log15.LvlDebug
	default:
		lvl, err := log15.LvlFromString(level)
		if err == nil {
			return lvl
		}
		return log15.LvlDebug
	}
}
