package testdata

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris-db/files"
	"github.com/eris-ltd/eris-db/server"

	stypes "github.com/eris-ltd/eris-db/manager/eris-mint/state/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/types"
)

const TendermintConfigDefault = `# This is a TOML config file.
# For more information, see https:/github.com/toml-lang/toml

moniker = "__MONIKER__"
seeds = ""
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = ""
rpc_laddr = ""
`

func CreateTempWorkDir(privValidator *types.PrivValidator, genesis *stypes.GenesisDoc, folderName string) (string, error) {

	workDir := path.Join(os.TempDir(), folderName)
	os.RemoveAll(workDir)
	errED := EnsureDir(workDir)

	if errED != nil {
		return "", errED
	}

	cfgName := path.Join(workDir, "config.toml")
	scName := path.Join(workDir, "server_conf.toml")
	pvName := path.Join(workDir, "priv_validator.json")
	genesisName := path.Join(workDir, "genesis.json")

	// Write config.
	errCFG := files.WriteFileRW(cfgName, []byte(TendermintConfigDefault))
	if errCFG != nil {
		return "", errCFG
	}
	fmt.Printf("File written: %s\n.", cfgName)

	// Write validator.
	errPV := writeJSON(pvName, privValidator)
	if errPV != nil {
		return "", errPV
	}
	fmt.Printf("File written: %s\n.", pvName)

	// Write genesis
	errG := writeJSON(genesisName, genesis)
	if errG != nil {
		return "", errG
	}
	fmt.Printf("File written: %s\n.", genesisName)

	// Write server config.
	errWC := server.WriteServerConfig(scName, server.DefaultServerConfig())
	if errWC != nil {
		return "", errWC
	}
	fmt.Printf("File written: %s\n.", scName)
	return workDir, nil
}

// Used to write json files using tendermints wire package.
func writeJSON(file string, v interface{}) error {
	var n int64
	var errW error
	fo, errC := os.Create(file)
	if errC != nil {
		return errC
	}
	wire.WriteJSON(v, fo, &n, &errW)
	if errW != nil {
		return errW
	}
	errL := fo.Close()
	if errL != nil {
		return errL
	}
	return nil
}
