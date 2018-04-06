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
)

// Monax-Keys server connects over http request-response structures

type HTTPResponse struct {
	Response string
	Error    string
}

type Requester func(method string, args map[string]string) (response string, err error)

func DefaultRequester(rpcAddress string, logger *logging.Logger) Requester {
	return func(method string, args map[string]string) (string, error) {
		body, err := json.Marshal(args)
		if err != nil {
			return "", err
		}
		endpoint := fmt.Sprintf("%s/%s", rpcAddress, method)
		logger.TraceMsg("Sending request to key server",
			"key_server_endpoint", endpoint,
			"request_body", string(body),
		)
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
		if err != nil {
			return "", err
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := requestResponse(req)
		if err != nil {
			return "", fmt.Errorf("error calling monax-keys at %s: %s", endpoint, err.Error())
		}
		if res.Error != "" {
			return "", fmt.Errorf("response error when calling monax-keys at %s: %s", endpoint, res.Error)
		}
		logger.TraceMsg("Received response from key server",
			"endpoint", endpoint,
			"request_body", string(body),
			"response", res,
		)
		return res.Response, nil
	}
}

func requestResponse(req *http.Request) (*HTTPResponse, error) {
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(resp.Status)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	httpResponse := new(HTTPResponse)
	if err := json.Unmarshal(bs, httpResponse); err != nil {
		return nil, err
	}
	return httpResponse, nil
}
