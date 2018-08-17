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

package exec

import (
	"strings"

	"fmt"

	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/tmthrgd/go-hex"
)

const logNTextTopicCutset = "\x00"
const LogNKeyPrefix = "Log"

func LogNKey(topic int) string {
	return fmt.Sprintf("%s%d", LogNKeyPrefix, topic)
}

func LogNTextKey(topic int) string {
	return fmt.Sprintf("%s%dText", LogNKeyPrefix, topic)
}

var logTagKeys []string
var logNTopicIndex = make(map[string]int, 5)
var logNTextTopicIndex = make(map[string]int, 5)

func init() {
	for i := 0; i <= 4; i++ {
		logN := LogNKey(i)
		logTagKeys = append(logTagKeys, LogNKey(i))
		logNText := LogNTextKey(i)
		logTagKeys = append(logTagKeys, logNText)
		logNTopicIndex[logN] = i
		logNTextTopicIndex[logNText] = i
	}
	logTagKeys = append(logTagKeys, event.AddressKey)
}

func (log *LogEvent) Get(key string) (string, bool) {
	if log == nil {
		return "", false
	}
	var value interface{}
	switch key {
	case event.AddressKey:
		value = log.Address
	default:
		if i, ok := logNTopicIndex[key]; ok {
			return hex.EncodeUpperToString(log.GetTopic(i).Bytes()), true
		}
		if i, ok := logNTextTopicIndex[key]; ok {
			return strings.Trim(string(log.GetTopic(i).Bytes()), logNTextTopicCutset), true
		}
		return "", false
	}
	return query.StringFromValue(value), true
}

func (log *LogEvent) GetTopic(i int) Word256 {
	if i < len(log.Topics) {
		return log.Topics[i]
	}
	return Word256{}
}

func (log *LogEvent) Len() int {
	return len(logTagKeys)
}

func (log *LogEvent) Keys() []string {
	return logTagKeys
}
