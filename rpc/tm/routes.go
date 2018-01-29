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

package tm

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm/method"
	"github.com/hyperledger/burrow/txs"
	gorpc "github.com/tendermint/tendermint/rpc/lib/server"
	"github.com/tendermint/tendermint/rpc/lib/types"
)

func GetRoutes(service rpc.Service) map[string]*gorpc.RPCFunc {
	return map[string]*gorpc.RPCFunc{
		// Transact
		method.BroadcastTx: gorpc.NewRPCFunc(func(tx txs.Wrapper) (rpc.Result, error) {
			return wrapReturnBurrowResult(service.BroadcastTx(tx.Unwrap()))
		}, "tx"),
		// Events
		method.Subscribe: gorpc.NewWSRPCFunc(func(wsCtx rpctypes.WSRPCContext,
			eventId string) (rpc.Result, error) {
			return wrapReturnBurrowResult(service.Subscribe(eventId,
				func(eventData event.AnyEventData) {
					// NOTE: EventSwitch callbacks must be nonblocking
					wsCtx.TryWriteRPCResponse(rpctypes.NewRPCSuccessResponse(wsCtx.Request.ID+"#event",
						rpc.ResultEvent{Event: eventId, AnyEventData: eventData}.Wrap()))
				}))
		}, "eventId"),
		method.Unsubscribe: gorpc.NewWSRPCFunc(func(wsCtx rpctypes.WSRPCContext,
			subscriptionId string) (rpc.Result, error) {
			return wrapReturnBurrowResult(service.Unsubscribe(subscriptionId))
		}, "subscriptionId"),
		// Status
		method.Status:  gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.Status), ""),
		method.NetInfo: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.NetInfo), ""),
		// Accounts
		method.ListAccounts: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.ListAccounts), ""),
		method.GetAccount:   gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.GetAccount), "address"),
		method.GetStorage:   gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.GetStorage), "address,key"),
		method.DumpStorage:  gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.DumpStorage), "address"),
		// Simulated call
		method.Call:     gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.Call), "fromAddress,toAddress,data"),
		method.CallCode: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.CallCode), "fromAddress,code,data"),
		// Blockchain
		method.Genesis:    gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.Genesis), ""),
		method.ChainID:    gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.ChainId), ""),
		method.Blockchain: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.BlockchainInfo), "minHeight,maxHeight"),
		method.GetBlock:   gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.GetBlock), "height"),
		// Consensus
		method.ListUnconfirmedTxs: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.ListUnconfirmedTxs), "maxTxs"),
		method.ListValidators:     gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.ListValidators), ""),
		method.DumpConsensusState: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.DumpConsensusState), ""),
		// Names
		method.GetName:   gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.GetName), "name"),
		method.ListNames: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.ListNames), ""),
		// Private keys and signing
		method.GeneratePrivateAccount: gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.GeneratePrivateAccount), ""),
		method.SignTx:                 gorpc.NewRPCFunc(mustWrapFuncBurrowResult(service.SignTx), "tx,privAccounts"),
	}
}

func mustWrapFuncBurrowResult(f interface{}) interface{} {
	wrapped, err := wrapFuncBurrowResult(f)
	if err != nil {
		panic(fmt.Errorf("must be able to wrap RPC function: %v", err))
	}
	return wrapped
}

// Takes a function with a covariant return type in func(args...) (ResultInner, error)
// and returns it as func(args...) (Result, error) so it can be serialised with the mapper
func wrapFuncBurrowResult(f interface{}) (interface{}, error) {
	rv := reflect.ValueOf(f)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		return nil, fmt.Errorf("must be passed a func f, but got: %#v", f)
	}

	in := make([]reflect.Type, rt.NumIn())
	for i := 0; i < rt.NumIn(); i++ {
		in[i] = rt.In(i)
	}

	if rt.NumOut() != 2 {
		return nil, fmt.Errorf("expects f to return the pair of ResultInner, error but got %v return types",
			rt.NumOut())
	}

	out := make([]reflect.Type, 2)
	err := checkTypeImplements(rt.Out(0), (*rpc.ResultInner)(nil))
	if err != nil {
		return nil, fmt.Errorf("wrong first return type: %v", err)
	}
	err = checkTypeImplements(rt.Out(1), (*error)(nil))
	if err != nil {
		return nil, fmt.Errorf("wrong second return type: %v", err)
	}

	out[0] = reflect.TypeOf(rpc.Result{})
	out[1] = rt.Out(1)

	return reflect.MakeFunc(reflect.FuncOf(in, out, false),
		func(args []reflect.Value) []reflect.Value {
			ret := rv.Call(args)
			burrowResult := reflect.New(out[0])
			burrowResult.Elem().Field(0).Set(ret[0])
			ret[0] = burrowResult.Elem()
			return ret
		}).Interface(), nil
}

func wrapReturnBurrowResult(result rpc.ResultInner, err error) (rpc.Result, error) {
	return rpc.Result{ResultInner: result}, err
}

// Passed a type and a pointer to an interface value checks that typ implements that interface
// returning a human-readable error if it does not
// (e.g. ifacePtr := (*MyInterface)(nil) can be a reasonably convenient stand-in for an actual type literal),
func checkTypeImplements(typ reflect.Type, ifacePtr interface{}) error {
	ifaceType := reflect.TypeOf(ifacePtr).Elem()
	if !typ.Implements(ifaceType) {
		return fmt.Errorf("%s does not implement interface %s", typ, ifaceType)
	}
	return nil
}
