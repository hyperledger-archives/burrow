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

// version provides the current Eris-DB version and a VersionIdentifier
// for the modules to identify their version with.

// This file is originally based on github.com/tendermint/tmsp/client/...
// .../local_client.go

package tendermint

import (
	"sync"

	tmsp_client "github.com/tendermint/tmsp/client"
	tmsp_types  "github.com/tendermint/tmsp/types"

	manager_types "github.com/eris-ltd/eris-db/manager/types"
)

// NOTE [ben] Compiler check to ensure localClient successfully implements
// tendermint/tmsp/client
var _ tmsp_client.Client = (*localClient)(nil)

type localClient struct {
	mtx         *sync.Mutex
	Application manager_types.Application
	Callback    tmsp_client.Callback
}

func NewLocalClient(mtx *sync.Mutex, app manager_types.Application) *localClient {
	if mtx == nil {
		mtx = new(sync.Mutex)
	}
	return &localClient{
		mtx:         mtx,
		Application: app,
	}
}

func (app *localClient) SetResponseCallback(cb tmsp_client.Callback) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	app.Callback = cb
}

// TODO: change manager_types.Application to include Error()?
func (app *localClient) Error() error {
	return nil
}

func (app *localClient) Stop() bool {
	return true
}

func (app *localClient) FlushAsync() *tmsp_client.ReqRes {
	// Do nothing
	return newLocalReqRes(tmsp_types.RequestFlush(), nil)
}

func (app *localClient) EchoAsync(msg string) *tmsp_client.ReqRes {
	return app.callback(
		tmsp_types.RequestEcho(msg),
		tmsp_types.ResponseEcho(msg),
	)
}

func (app *localClient) InfoAsync() *tmsp_client.ReqRes {
	app.mtx.Lock()
	info := app.Application.Info()
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestInfo(),
		tmsp_types.ResponseInfo(info),
	)
}

func (app *localClient) SetOptionAsync(key string, value string) *tmsp_client.ReqRes {
	app.mtx.Lock()
	log := app.Application.SetOption(key, value)
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestSetOption(key, value),
		tmsp_types.ResponseSetOption(log),
	)
}

func (app *localClient) AppendTxAsync(tx []byte) *tmsp_client.ReqRes {
	app.mtx.Lock()
	res := app.Application.AppendTx(tx)
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestAppendTx(tx),
		tmsp_types.ResponseAppendTx(res.Code, res.Data, res.Log),
	)
}

func (app *localClient) CheckTxAsync(tx []byte) *tmsp_client.ReqRes {
	app.mtx.Lock()
	res := app.Application.CheckTx(tx)
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestCheckTx(tx),
		tmsp_types.ResponseCheckTx(res.Code, res.Data, res.Log),
	)
}

func (app *localClient) QueryAsync(tx []byte) *tmsp_client.ReqRes {
	app.mtx.Lock()
	res := app.Application.Query(tx)
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestQuery(tx),
		tmsp_types.ResponseQuery(res.Code, res.Data, res.Log),
	)
}

func (app *localClient) CommitAsync() *tmsp_client.ReqRes {
	app.mtx.Lock()
	res := app.Application.Commit()
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestCommit(),
		tmsp_types.ResponseCommit(res.Code, res.Data, res.Log),
	)
}

func (app *localClient) InitChainAsync(validators []*tmsp_types.Validator) *tmsp_client.ReqRes {
	app.mtx.Lock()
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		bcApp.InitChain(validators)
	}
	reqRes := app.callback(
		tmsp_types.RequestInitChain(validators),
		tmsp_types.ResponseInitChain(),
	)
	app.mtx.Unlock()
	return reqRes
}

func (app *localClient) BeginBlockAsync(height uint64) *tmsp_client.ReqRes {
	app.mtx.Lock()
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		bcApp.BeginBlock(height)
	}
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestBeginBlock(height),
		tmsp_types.ResponseBeginBlock(),
	)
}

func (app *localClient) EndBlockAsync(height uint64) *tmsp_client.ReqRes {
	app.mtx.Lock()
	var validators []*tmsp_types.Validator
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		validators = bcApp.EndBlock(height)
	}
	app.mtx.Unlock()
	return app.callback(
		tmsp_types.RequestEndBlock(height),
		tmsp_types.ResponseEndBlock(validators),
	)
}

//-------------------------------------------------------

func (app *localClient) FlushSync() error {
	return nil
}

func (app *localClient) EchoSync(msg string) (res tmsp_types.Result) {
	return tmsp_types.OK.SetData([]byte(msg))
}

func (app *localClient) InfoSync() (res tmsp_types.Result) {
	app.mtx.Lock()
	info := app.Application.Info()
	app.mtx.Unlock()
	return tmsp_types.OK.SetData([]byte(info))
}

func (app *localClient) SetOptionSync(key string, value string) (res tmsp_types.Result) {
	app.mtx.Lock()
	log := app.Application.SetOption(key, value)
	app.mtx.Unlock()
	return tmsp_types.OK.SetLog(log)
}

func (app *localClient) AppendTxSync(tx []byte) (res tmsp_types.Result) {
	app.mtx.Lock()
	res = app.Application.AppendTx(tx)
	app.mtx.Unlock()
	return res
}

func (app *localClient) CheckTxSync(tx []byte) (res tmsp_types.Result) {
	app.mtx.Lock()
	res = app.Application.CheckTx(tx)
	app.mtx.Unlock()
	return res
}

func (app *localClient) QuerySync(query []byte) (res tmsp_types.Result) {
	app.mtx.Lock()
	res = app.Application.Query(query)
	app.mtx.Unlock()
	return res
}

func (app *localClient) CommitSync() (res tmsp_types.Result) {
	app.mtx.Lock()
	res = app.Application.Commit()
	app.mtx.Unlock()
	return res
}

func (app *localClient) InitChainSync(validators []*tmsp_types.Validator) (err error) {
	app.mtx.Lock()
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		bcApp.InitChain(validators)
	}
	app.mtx.Unlock()
	return nil
}

func (app *localClient) BeginBlockSync(height uint64) (err error) {
	app.mtx.Lock()
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		bcApp.BeginBlock(height)
	}
	app.mtx.Unlock()
	return nil
}

func (app *localClient) EndBlockSync(height uint64) (changedValidators []*tmsp_types.Validator, err error) {
	app.mtx.Lock()
	if bcApp, ok := app.Application.(tmsp_types.BlockchainAware); ok {
		changedValidators = bcApp.EndBlock(height)
	}
	app.mtx.Unlock()
	return changedValidators, nil
}

//-------------------------------------------------------

func (app *localClient) callback(req *tmsp_types.Request, res *tmsp_types.Response) *tmsp_client.ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}

func newLocalReqRes(req *tmsp_types.Request, res *tmsp_types.Response) *tmsp_client.ReqRes {
	reqRes := tmsp_client.NewReqRes(req)
	reqRes.Response = res
	reqRes.SetDone()
	return reqRes
}
