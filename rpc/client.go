package rpc

type Client interface {
	Call(method string, params interface{}, result interface{}) error
}
