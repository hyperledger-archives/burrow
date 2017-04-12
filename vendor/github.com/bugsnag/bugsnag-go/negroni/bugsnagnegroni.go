package bugsnagnegroni

import (
	"github.com/bugsnag/bugsnag-go"
	"net/http"
)

type handler struct {
	rawData []interface{}
}

func AutoNotify(rawData ...interface{}) *handler {
	return &handler{
		rawData: rawData,
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	notifier := bugsnag.New(append(h.rawData, r)...)
	defer notifier.AutoNotify(r)
	next(rw, r)
}
