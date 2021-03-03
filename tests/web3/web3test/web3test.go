package web3test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/rpc/lib/jsonrpc"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
)

const ganacheRemote = "http://127.0.0.1:8545"

// Configured in the "ganache" yarn script in vent/test/eth/package.json
var ganachePrivateKey = hex.MustDecodeString("cfad15e9e8f24b5b5608cac150293fe971d23cd9168206b231c859b7974d4295")

// Set INFURA_SECRET to the basic auth password to use infura remote
var infuraRemote = fmt.Sprintf("https://:%s@ropsten.infura.io/v3/7ed3059377654803a190fa44560d528f",
	os.Getenv("INFURA_SECRET"))

// Toggle below to switch to an infura test
var remote = ganacheRemote

//var remote = infuraRemote

var client = jsonrpc.NewClient(remote)

func GetChainRemote() string {
	return remote
}

func GetChainRPCClient() *jsonrpc.Client {
	return client
}

func GetPrivateKey(t testing.TB) *crypto.PrivateKey {
	if remote == ganacheRemote {
		pk, err := crypto.PrivateKeyFromRawBytes(ganachePrivateKey, crypto.CurveTypeSecp256k1)
		require.NoError(t, err)
		return &pk
	}
	// This account (5DA093B66C2D373E4CBB6081312BE5DFCFF66189) had some test ether on Ropsten at some point:
	// https://ropsten.etherscan.io/address/0x5DA093B66C2D373E4CBB6081312BE5DFCFF66189
	// https://faucet.dimensions.network/ seems to be less trigger happy on banning you to top up eth supply
	pk := crypto.PrivateKeyFromSecret("fooooo", crypto.CurveTypeSecp256k1)
	fmt.Println(pk.GetAddress())
	return &pk
}
