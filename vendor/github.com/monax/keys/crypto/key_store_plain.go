/*
	This file is part of go-ethereum

	go-ethereum is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	go-ethereum is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with go-ethereum.  If not, see <http://www.gnu.org/licenses/>.
*/
/**
 * @authors
 * 	Gustav Simonsson <gustav.simonsson@gmail.com>
 * 	Ethan Buchman <ethan@erisindustries.com> (slight modifications)
 * @date 2015
 *

 */

package crypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type KeyStore interface {
	GenerateNewKey(tpy KeyType, auth string) (*Key, error)
	GetKey(addr []byte, auth string) (*Key, error)
	GetAllAddresses() ([][]byte, error)
	StoreKey(key *Key, auth string) error
	DeleteKey(addr []byte, auth string) error
}

type keyStorePlain struct {
	keysDirPath string
}

func NewKeyStorePlain(path string) KeyStore {
	return &keyStorePlain{path}
}

func (ks keyStorePlain) GenerateNewKey(typ KeyType, auth string) (key *Key, err error) {
	return GenerateNewKeyDefault(ks, typ, auth)
}

func GenerateNewKeyDefault(ks KeyStore, typ KeyType, auth string) (key *Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("GenerateNewKey error: %v", r)
		}
	}()
	key, err = NewKey(typ)
	if err != nil {
		return nil, err
	}
	err = ks.StoreKey(key, auth)
	return key, err
}

func (ks keyStorePlain) GetKey(keyAddr []byte, auth string) (key *Key, err error) {
	fileContent, err := GetKeyFile(ks.keysDirPath, keyAddr)
	if err != nil {
		return nil, err
	}

	key = new(Key)
	err = key.UnmarshalJSON(fileContent)
	return key, err
}

func (ks keyStorePlain) GetAllAddresses() (addresses [][]byte, err error) {
	return GetAllAddresses(ks.keysDirPath)
}

func (ks keyStorePlain) StoreKey(key *Key, auth string) (err error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return err
	}
	err = WriteKeyFile(key.Address, ks.keysDirPath, keyJSON)
	return err
}

func (ks keyStorePlain) DeleteKey(keyAddr []byte, auth string) (err error) {
	keyDirPath := path.Join(ks.keysDirPath, strings.ToUpper(hex.EncodeToString(keyAddr)))
	err = os.RemoveAll(keyDirPath)
	return err
}

func GetKeyFile(keysDirPath string, keyAddr []byte) (fileContent []byte, err error) {
	fileName := strings.ToUpper(hex.EncodeToString(keyAddr))
	return ioutil.ReadFile(path.Join(keysDirPath, fileName, fileName))
}

func WriteKeyFile(addr []byte, keysDirPath string, content []byte) (err error) {
	addrHex := strings.ToUpper(hex.EncodeToString(addr))
	keyDirPath := path.Join(keysDirPath, addrHex)
	keyFilePath := path.Join(keyDirPath, addrHex)
	err = os.MkdirAll(keyDirPath, 0700) // read, write and dir search for user
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keyFilePath, content, 0600) // read, write for user
}

func GetAllAddresses(keysDirPath string) (addresses [][]byte, err error) {
	fileInfos, err := ioutil.ReadDir(keysDirPath)
	if err != nil {
		return nil, err
	}
	addresses = make([][]byte, len(fileInfos))
	for i, fileInfo := range fileInfos {
		address, err := hex.DecodeString(fileInfo.Name())
		if err != nil {
			continue
		}
		addresses[i] = address
	}
	return addresses, err
}
