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

package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
)

// Monax-Keys server connects over http request-response structures

type HTTPResponse struct {
	Response string
	Error    string
}

func RequestResponse(addr, method string, args map[string]string, logger logging_types.InfoTraceLogger) (string, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s/%s", addr, method)
	logging.TraceMsg(logger, "Sending request to key server",
		"key_server_endpoint", endpoint,
		"request_body", string(body),
	)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, errS, err := requestResponse(req)
	if err != nil {
		return "", fmt.Errorf("Error calling monax-keys at %s: %s", endpoint, err.Error())
	}
	if errS != "" {
		return "", fmt.Errorf("Error (string) calling monax-keys at %s: %s", endpoint, errS)
	}
	logging.TraceMsg(logger, "Received response from key server",
		"endpoint", endpoint,
		"request body", string(body),
		"response", res,
	)
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
