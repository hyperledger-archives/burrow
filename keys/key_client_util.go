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

package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/eris-ltd/eris-logger"
)

// Eris-Keys server connects over http request-response structures

type HTTPResponse struct {
	Response string
	Error    string
}

func RequestResponse(addr, method string, args map[string]string) (string, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s/%s", addr, method)
	log.WithFields(log.Fields{
		"key server endpoint": endpoint,
		"request body": string(b),
		}).Debugf("Sending request body to key server")
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, errS, err := requestResponse(req)
	if err != nil {
		return "", fmt.Errorf("Error calling eris-keys at %s: %s", endpoint, err.Error())
	}
	if errS != "" {
		return "", fmt.Errorf("Error (string) calling eris-keys at %s: %s", endpoint, errS)
	}
	log.WithFields(log.Fields{
		"endpoint": endpoint,
		"request body": string(b),
		"response": res,
		}).Debugf("Received response from key server")
	return res, nil
}

func requestResponse(req *http.Request) (string, string, error) {
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf(resp.Status)
	}
	return unpackResponse(resp)
}

func unpackResponse(resp *http.Response) (string, string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	r := new(HTTPResponse)
	if err := json.Unmarshal(b, r); err != nil {
		return "", "", err
	}
	return r.Response, r.Error, nil
}