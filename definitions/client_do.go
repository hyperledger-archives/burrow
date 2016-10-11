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
