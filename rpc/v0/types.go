package v0

type ClientVersion struct {
	ClientVersion string `json:"client_version"`
}

type Moniker struct {
	Moniker string `json:"moniker"`
}

type Listening struct {
	Listening bool `json:"listening"`
}

type Listeners struct {
	Listeners []string `json:"listeners"`
}
