package node

import (
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/config/tendermint_test"
	dbm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/db"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/p2p"
	sm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	stypes "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state/types"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

func TestNodeStartStop(t *testing.T) {
	// Create & start node
	n := NewNodeDefaultPrivVal()
	l := p2p.NewDefaultListener("tcp", config.GetString("node_laddr"))
	n.AddListener(l)
	n.Start()
	log.Notice("Started node", "nodeInfo", n.sw.NodeInfo())
	time.Sleep(time.Second * 2)
	ch := make(chan struct{}, 1)
	go func() {
		n.Stop()
		ch <- struct{}{}
	}()
	ticker := time.NewTicker(time.Second * 5)
	select {
	case <-ch:
	case <-ticker.C:
		t.Fatal("timed out waiting for shutdown")
	}
}

func TestCompatibleNodeInfo(t *testing.T) {
	sw := p2p.NewSwitch()
	priv1, priv2 := types.GenPrivValidator(), types.GenPrivValidator()
	genDoc1, _, _ := stypes.RandGenesisDoc(5, true, 100, 4, true, 1000)
	genState1 := sm.MakeGenesisState(dbm.NewMemDB(), genDoc1)
	genDoc2, _, _ := stypes.RandGenesisDoc(5, true, 100, 4, true, 1000)
	genState2 := sm.MakeGenesisState(dbm.NewMemDB(), genDoc2)

	// incompatible genesis states
	n1 := makeNodeInfo(sw, priv1.PrivKey, genState1.Hash())
	n2 := makeNodeInfo(sw, priv2.PrivKey, genState2.Hash())
	if err := n1.CompatibleWith(n2); err == nil {
		t.Fatalf("Expected nodes to be incompatible due to genesis state")
	}

	// incompatible chain ids
	copy(n2.Genesis, n1.Genesis)
	n2.ChainID = "incryptowetrust"
	if err := n1.CompatibleWith(n2); err == nil {
		t.Fatalf("Expected nodes to be incompatible due to chain ID")
	}

	// incompatible versions
	n2.ChainID = n1.ChainID
	v := n1.Version.Tendermint
	spl := strings.Split(v, ".")
	n, err := strconv.Atoi(spl[0])
	if err != nil {
		t.Fatalf(err.Error())
	}
	spl[0] = strconv.Itoa(n + 1)
	n2.Version.Tendermint = strings.Join(spl, ".")
	if err := n1.CompatibleWith(n2); err == nil {
		t.Fatalf("Expected nodes to be incompatible due to major version")
	}

	// compatible
	n2.Version.Tendermint = n1.Version.Tendermint
	if err := n1.CompatibleWith(n2); err != nil {
		t.Fatalf("Expected nodes to be compatible")
	}
}
