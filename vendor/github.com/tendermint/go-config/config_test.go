package config

import (
	"testing"

	"github.com/BurntSushi/toml"
	. "github.com/tendermint/go-common"
)

var (
	stringTable = map[string]string{
		"astring":                  "a",
		"subfield1.astring":        "b",
		"subfield1.subsub.astring": "b",
		"subfield2.astring":        "c",
	}

	intTable = map[string]int{
		"anint": 42,
	}

	boolTable = map[string]bool{
		"abool": false,
	}
)

var testTxt = Fmt(`

# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml


astring = "%s"
anint = %d
abool = %v

[subfield1]
	astring = "%s"
	[subfield1.subsub]
	astring = "%s"


[subfield2]
astring = "%s"
`, stringTable["astring"], intTable["anint"], boolTable["abool"], stringTable["subfield1.astring"],
	stringTable["subfield1.subsub.astring"],
	stringTable["subfield2.astring"])

func TestConfig(t *testing.T) {
	var configData = make(map[string]interface{})
	err := toml.Unmarshal([]byte(testTxt), &configData)
	if err != nil {
		t.Fatal(err)
	}

	cfg := NewMapConfig(configData)

	for k, v := range stringTable {
		if x := cfg.GetString(k); x != v {
			t.Fatalf("Got %v. Expected %v", x, v)
		}
	}

	for k, v := range intTable {
		if x := cfg.GetInt(k); x != v {
			t.Fatalf("Got %v. Expected %v", x, v)
		}
	}

	for k, v := range boolTable {
		if x := cfg.GetBool(k); x != v {
			t.Fatalf("Got %v. Expected %v", x, v)
		}
	}

}

func TestSetConfig(t *testing.T) {
	var configData = make(map[string]interface{})

	cfg := NewMapConfig(configData)
	cfg.Set("abc.def", "x")
	if x := cfg.GetString("abc.def"); x != "x" {
		t.Fatalf("Got %v, expected 1", x)
	}
}
