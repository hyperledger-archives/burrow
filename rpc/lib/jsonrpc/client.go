package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/hyperledger/burrow/rpc/lib/types"
	"github.com/pkg/errors"
)

// HTTPClient is a common interface for Client and URIClient.
type HTTPClient interface {
	Call(method string, params interface{}, result interface{}) error
}

// Client takes params as a slice
type Client struct {
	address string
	client  http.Client
}

// NewClient returns a Client pointed at the given address.
func NewClient(remote string) *Client {
	return &Client{
		address: remote,
	}
}

func (c *Client) Call(method string, params interface{}, result interface{}) error {
	request, err := types.NewRequest("jsonrpc-client", method, params)
	if err != nil {
		return err
	}
	bs, err := json.Marshal(request)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(bs)
	response, err := c.client.Post(c.address, "application/json", buf)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bs, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return unmarshalResponseBytes(bs, result)
}

// URI takes params as a map
type URIClient struct {
	address string
	client  http.Client
}

func NewURIClient(remote string) *URIClient {
	return &URIClient{
		address: remote,
	}
}

func (c *URIClient) Call(method string, params interface{}, result interface{}) error {
	values, err := argsToURLValues(params.(map[string]interface{}))
	if err != nil {
		return err
	}
	// log.Info(Fmt("URI request to %v (%v): %v", c.address, method, values))
	resp, err := c.client.PostForm(c.address+"/"+method, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint: errcheck

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return unmarshalResponseBytes(responseBytes, result)
}

func unmarshalResponseBytes(responseBytes []byte, result interface{}) error {
	var err error
	response := &types.RPCResponse{}
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return errors.Errorf("Error unmarshalling rpc response: %v", err)
	}
	if response.Error != nil {
		return response.Error
	}
	// Unmarshal the RawMessage into the result.
	err = json.Unmarshal(response.Result, result)
	if err != nil {
		return errors.Errorf("error unmarshalling rpc response result: %v", err)
	}
	return nil
}

func argsToURLValues(args map[string]interface{}) (url.Values, error) {
	values := make(url.Values)
	if len(args) == 0 {
		return values, nil
	}
	err := argsToJSON(args)
	if err != nil {
		return nil, err
	}
	for key, val := range args {
		values.Set(key, val.(string))
	}
	return values, nil
}

func argsToJSON(args map[string]interface{}) error {
	for k, v := range args {
		rt := reflect.TypeOf(v)
		isByteSlice := rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8
		if isByteSlice {
			bs := reflect.ValueOf(v).Bytes()
			args[k] = fmt.Sprintf("0x%X", bs)
			continue
		}

		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		args[k] = string(data)
	}
	return nil
}
