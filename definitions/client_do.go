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

type ClientDo {
	// Persistent flags not reflected in the configuration files
	// only set through command line flags or environment variables
	Debug   bool // ERIS_DB_DEBUG
	Verbose bool // ERIS_DB_VERBOSE

	// Following parameters are global flags for eris-client
	signAddrFlag string
	nodeAddrFlag string
	pubkeyFlag   string
	addrFlag     string
	chainidFlag  string

	// signFlag      bool // TODO: remove; unsafe signing without eris-keys
	broadcastFlag bool
	waitFlag      bool

	// Following parameters are specific for Transaction command
	// some of these are strings rather than flags because the `core`
	// functions have a pure string interface so they work nicely from http
	amtFlag      string
	nonceFlag    string
	nameFlag     string
	dataFlag     string
	dataFileFlag string
	toFlag       string
	feeFlag      string
	gasFlag      string
	unbondtoFlag string
	heightFlag   string
}

func NewClientDo() *ClientDo {
	clientDo := new(ClientDo)
	clientDo.Debug = false
	clientDo.Verbose = false
	
	clientDo.signAddrFlag = ""
	clientDo.nodeAddrFlag = ""
	clientDo.pubkeyFlag = ""
	clientDo.addrFlag = ""
	clientDo.chainidFlag = ""

	clientDo.signFlag = false
	clientDo.broadcastFlag = false
	clientDo.waitFlag = false

	clientDo.amtFlag = ""
	clientDo.nonceFlag = ""
	clientDo.nameFlag = ""
	clientDo.dataFlag = ""
	clientDo.dataFileFlag = ""
	clientDo.toFlag = ""
	clientDo.feeFlag = ""
	clientDo.gasFlag = ""
	clientDo.unbondtoFlag = ""
	clientDo.heightFlag = ""

	return clientDo
}