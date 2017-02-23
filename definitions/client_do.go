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

package definitions

type ClientDo struct {
	// Persistent flags not reflected in the configuration files
	// only set through command line flags or environment variables
	Debug   bool // ERIS_DB_DEBUG
	Verbose bool // ERIS_DB_VERBOSE

	// Following parameters are global flags for eris-client tx
	SignAddrFlag string
	NodeAddrFlag string
	PubkeyFlag   string
	AddrFlag     string
	ChainidFlag  string

	// signFlag      bool // TODO: remove; unsafe signing without eris-keys
	BroadcastFlag bool
	WaitFlag      bool

	// Following parameters are vary for different Transaction subcommands
	// some of these are strings rather than flags because the `core`
	// functions have a pure string interface so they work nicely from http
	AmtFlag      string
	NonceFlag    string
	NameFlag     string
	DataFlag     string
	DataFileFlag string
	ToFlag       string
	FeeFlag      string
	GasFlag      string
	UnbondtoFlag string
	HeightFlag   string
}

func NewClientDo() *ClientDo {
	clientDo := new(ClientDo)
	clientDo.Debug = false
	clientDo.Verbose = false

	clientDo.SignAddrFlag = ""
	clientDo.NodeAddrFlag = ""
	clientDo.PubkeyFlag = ""
	clientDo.AddrFlag = ""
	clientDo.ChainidFlag = ""

	// clientDo.signFlag = false
	clientDo.BroadcastFlag = false
	clientDo.WaitFlag = false

	clientDo.AmtFlag = ""
	clientDo.NonceFlag = ""
	clientDo.NameFlag = ""
	clientDo.DataFlag = ""
	clientDo.DataFileFlag = ""
	clientDo.ToFlag = ""
	clientDo.FeeFlag = ""
	clientDo.GasFlag = ""
	clientDo.UnbondtoFlag = ""
	clientDo.HeightFlag = ""

	return clientDo
}
