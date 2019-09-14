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

package native

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
)

type Natives struct {
	callableByAddress map[crypto.Address]Native
	callableByName    map[string]Native
	logger            *logging.Logger
}

func New() *Natives {
	return &Natives{
		callableByAddress: make(map[crypto.Address]Native),
		callableByName:    make(map[string]Native),
		logger:            logging.NewNoopLogger(),
	}
}

func Merge(nss ...*Natives) (*Natives, error) {
	n := New()
	for _, ns := range nss {
		for _, contract := range ns.callableByName {
			err := n.register(contract)
			if err != nil {
				return nil, err
			}
		}
	}
	return n, nil
}

func (ns *Natives) WithLogger(logger *logging.Logger) *Natives {
	ns.logger = logger
	return ns
}

func (ns *Natives) Dispatch(acc *acm.Account) engine.Callable {
	return ns.GetByAddress(acc.Address)
}

func (ns *Natives) SetExternals(externals engine.Dispatcher) {
	// TODO: thread these through to the underlying functions
}

func (ns *Natives) Callables() []engine.Callable {
	callables := make([]engine.Callable, 0, len(ns.callableByAddress))
	for _, c := range ns.callableByAddress {
		callables = append(callables, c)
	}
	return callables
}

func (ns *Natives) GetByName(name string) Native {
	return ns.callableByName[name]
}

func (ns *Natives) GetContract(name string) *Contract {
	c, _ := ns.GetByName(name).(*Contract)
	return c
}

func (ns *Natives) GetFunction(name string) *Function {
	f, _ := ns.GetByName(name).(*Function)
	return f
}

func (ns *Natives) GetByAddress(address crypto.Address) Native {
	return ns.callableByAddress[address]
}

func (ns *Natives) IsRegistered(address crypto.Address) bool {
	_, ok := ns.callableByAddress[address]
	return ok
}

func (ns *Natives) MustContract(name, comment string, functions ...Function) *Natives {
	ns, err := ns.Contract(name, comment, functions...)
	if err != nil {
		panic(err)
	}
	return ns
}

func (ns *Natives) Contract(name, comment string, functions ...Function) (*Natives, error) {
	contract, err := NewContract(name, comment, ns.logger, functions...)
	if err != nil {
		return nil, err
	}
	err = ns.register(contract)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (ns *Natives) MustFunction(comment string, address crypto.Address, permFlag permission.PermFlag, f interface{}) *Natives {
	ns, err := ns.Function(comment, address, permFlag, f)
	if err != nil {
		panic(err)
	}
	return ns
}

func (ns *Natives) Function(comment string, address crypto.Address, permFlag permission.PermFlag, f interface{}) (*Natives, error) {
	function, err := NewFunction(comment, address, permFlag, f)
	if err != nil {
		return nil, err
	}
	err = ns.register(function)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (ns *Natives) register(callable Native) error {
	name := callable.FullName()
	address := callable.Address()
	_, ok := ns.callableByName[name]
	if ok {
		return fmt.Errorf("cannot redeclare contract with name %s", name)
	}
	_, ok = ns.callableByAddress[address]
	if ok {
		return fmt.Errorf("cannot redeclare contract with address %v", address)
	}
	ns.callableByName[name] = callable
	ns.callableByAddress[address] = callable
	return nil
}
