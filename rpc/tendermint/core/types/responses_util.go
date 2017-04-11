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

package types

import (
	"github.com/hyperledger/burrow/core/types"
)

// UnwrapResultDumpStorage does a deep copy to remove /rpc/tendermint/core/types
// and expose /core/types instead.  This is largely an artefact to be removed once
// go-wire and go-rpc are deprecated.
// This is not an efficient code, especially given Storage can be big.
func UnwrapResultDumpStorage(result *ResultDumpStorage) *types.Storage {
	storageRoot := make([]byte, len(result.StorageRoot))
	copy(storageRoot, result.StorageRoot)
	storageItems := make([]types.StorageItem, len(result.StorageItems))
	for i, item := range result.StorageItems {
		key := make([]byte, len(item.Key))
		value := make([]byte, len(item.Value))
		copy(key, item.Key)
		copy(value, item.Value)
		storageItems[i] = types.StorageItem{
			Key:   key,
			Value: value,
		}
	}
	return &types.Storage{
		StorageRoot:  storageRoot,
		StorageItems: storageItems,
	}
}
