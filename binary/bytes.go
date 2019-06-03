// Copyright 2019 Monax Industries Limited
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

package binary

import hex "github.com/tmthrgd/go-hex"

type HexBytes []byte

func (hb *HexBytes) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	*hb = bs
	return nil
}

func (hb HexBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

func (hb HexBytes) String() string {
	return hex.EncodeUpperToString(hb)
}

// Protobuf support
func (hb HexBytes) Marshal() ([]byte, error) {
	return hb, nil
}

func (hb *HexBytes) Unmarshal(data []byte) error {
	*hb = data
	return nil
}

func (hb HexBytes) MarshalTo(data []byte) (int, error) {
	return copy(data, hb), nil
}

func (hb HexBytes) Size() int {
	return len(hb)
}

func (hb HexBytes) Bytes() []byte {
	return hb
}
