// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package core_types

import (
	"github.com/eris-ltd/eris-db/core/types"
)

// UnwrapResultDumpStorage does a deep copy to remove /rpc/tendermint/core/types
// and expose /core/types instead.  This is largely an artefact to be removed once
// go-wire and go-rpc are deprecated.
// This is not an efficient code, especially given Storage can be big.
func UnwrapResultDumpStorage(result *ResultDumpStorage) (*types.Storage) {
	storageRoot := make([]byte, len(result.StorageRoot))
	copy(storageRoot, result.StorageRoot)
	storageItems := make([]types.StorageItem, len(result.StorageItems))
	for i, item := range result.StorageItems {
		key := make([]byte, len(item.Key))
		value := make([]byte, len(item.Value))
		copy(key,  item.Key)
		copy(value, item.Value)
		storageItems[i] = types.StorageItem{
			Key: key,
			Value: value,
		}
	}
	return &types.Storage{
		StorageRoot: storageRoot,
		StorageItems: storageItems,
	}
}
