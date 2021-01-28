package rlp

import (
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/crypto"

	"github.com/test-go/testify/require"
)

type testCase struct {
	in  interface{}
	enc []byte
	dec interface{}
}

type testObject struct {
	Key   string
	Value string
}

func TestEncoding(t *testing.T) {

	t.Run("Empty", func(t *testing.T) {
		var tests = []testCase{
			{
				[]byte{},
				[]byte{uint8(StringOffset)},
				[]byte{},
			},
			{
				"",
				[]byte{uint8(StringOffset)},
				[]byte{},
			},
			{
				0,
				[]byte{uint8(StringOffset)},
				[]byte{},
			},
			{
				[]string{},
				[]byte{uint8(SliceOffset)},
				[]byte{},
			},
		}

		trial(t, tests)
	})

	t.Run("Bool", func(t *testing.T) {
		var tests = []testCase{
			{
				true,
				[]byte{0x01},
				[]byte{1},
			},
			{
				false,
				[]byte{uint8(StringOffset)},
				[]byte{0},
			},
		}

		trial(t, tests)
	})

	t.Run("String", func(t *testing.T) {
		var tests = []testCase{
			{
				[]byte{0, 0},
				[]byte{uint8(StringOffset) + 2, 0, 0},
				[]byte{0, 0},
			},
			{
				[]byte{0x64, 0x6f, 0x67},
				[]byte{0x83, 100, 111, 103},
				[]byte{0x64, 0x6f, 0x67},
			},
			{
				"dog",
				[]byte{0x83, 100, 111, 103},
				[]byte("dog"),
			},
			{
				"hello world",
				[]byte{0x8b, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64},
				[]byte("hello world"),
			},
			{
				"Lorem ipsum dolor sit amet, consectetur adipisicing elit",
				[]byte{0xb8, 0x38, 0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6d, 0x20, 0x64, 0x6f, 0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x63, 0x74, 0x65, 0x74, 0x75, 0x72, 0x20, 0x61, 0x64, 0x69, 0x70, 0x69, 0x73, 0x69, 0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6c, 0x69, 0x74},
				[]byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit"),
			},
			{
				[]byte{0x0f},
				[]byte{0x0f},
				[]byte{0x0f},
			},
			{
				[]byte{0x04, 0x00},
				[]byte{0x82, 0x04, 0x00},
				[]byte{0x04, 0x00},
			},
		}

		trial(t, tests)
	})

	t.Run("List", func(t *testing.T) {
		var tests = []testCase{
			{
				[]string{"cat", "dog"},
				[]byte{0xc8, 0x83, byte('c'), byte('a'), byte('t'), 0x83, byte('d'), byte('o'), byte('g')},
				[][]byte{[]byte("cat"), []byte("dog")},
			},
			{
				[][]string{[]string{"cat", "dog"}, []string{"owl"}},
				[]byte{0xce, 0xc8, 0x83, byte('c'), byte('a'), byte('t'), 0x83, byte('d'), byte('o'), byte('g'), 0xc4, 0x83, byte('o'), byte('w'), byte('l')},
				[][]byte{[]byte("cat"), []byte("dog"), []byte("owl")},
			},
		}

		trial(t, tests)
	})

	t.Run("Struct", func(t *testing.T) {
		var tests = []testCase{
			{
				testObject{"foo", "bar"},
				[]byte{0xc8, 0x83, byte('f'), byte('o'), byte('o'), 0x83, byte('b'), byte('a'), byte('r')},
				&testObject{"foo", "bar"},
			},
		}

		trial(t, tests)
	})
}

func trial(t *testing.T, tests []testCase) {
	for _, tt := range tests {
		enc, err := Encode(tt.in)
		require.NoError(t, err)
		require.Equal(t, tt.enc, enc, "encoding must match")

		var dec interface{}

		switch todo := tt.dec.(type) {
		case []byte:
			dec = make([]byte, len(todo))
		case [][]byte:
			dec = make([][]byte, len(todo))
		case *testObject:
			dec = new(testObject)
		default:
			require.FailNow(t, "dec type unsupported")
		}

		err = Decode(enc, dec)
		require.NoError(t, err)
		require.Equal(t, tt.dec, dec, "decoding must match")
	}
}

type RawTx struct {
	Nonce    uint64 `json:"nonce"`
	GasPrice uint64 `json:"gasPrice"`
	Gas      uint64 `json:"gas"`
	To       []byte `json:"to"`
	Value    uint64 `json:"value"`
	Input    []byte `json:"input"`

	V *big.Int `json:"v"`
	R []byte   `json:"r"`
	S []byte   `json:"s"`
}

func TestEthTransaction(t *testing.T) {
	// raw := `f866068609184e72a0008303000094fa3caabc8eefec2b5e2895e5afbf79379e7268a7808025a06d35f407f418737eec80cba738c4301e683cfcecf19bac9a1aeb2316cac19d3ba002935ee46e3b6bd69168b0b07670699d71df5b32d5f66dbca5758bce2431c9e8`
	// data, err := hex.DecodeString(raw)
	// require.NoError(t, err)

	//input := []interface{}{
	//	uint64(6),              // Nonce
	//	uint64(10000000000000), // GasPrice
	//	uint64(196608),         // Gas
	//	[]byte{250, 60, 170, 188, 142, 239, 236, 43, 94, 40, 149, 229, 175, 191, 121, 55, 158, 114, 104, 167}, // To
	//	uint64(0), // Value
	//	[]byte{},  // Input
	//	uint64(1), // V
	//	uint(0),   // R
	//	uint(0),   // S
	//}
	input := &RawTx{
		uint64(6),              // Nonce
		uint64(10000000000000), // GasPrice
		uint64(196608),         // Gas
		[]byte{250, 60, 170, 188, 142, 239, 236, 43, 94, 40, 149, 229, 175, 191, 121, 55, 158, 114, 104, 167}, // To
		uint64(0),     // Value
		[]byte{},      // Input
		big.NewInt(1), // V
		[]byte{1},     // R
		[]byte{1},     // S
	}
	data, err := Encode(input)
	require.NoError(t, err)

	exp := []byte{230, 6, 134, 9, 24, 78, 114, 160, 0, 131, 3, 0, 0, 148, 250, 60, 170, 188, 142, 239, 236, 43, 94, 40, 149, 229, 175, 191, 121, 55, 158, 114, 104, 167, 128, 128, 1, 1, 1}
	require.Equal(t, exp, data)

	tx := new(RawTx)
	err = Decode(data, tx)
	require.NoError(t, err)

	require.Equal(t, input, tx)
}

func TestBigInts(t *testing.T) {
	type foo struct {
		A *big.Int
		B *big.Int
	}

	biggun, ok := new(big.Int).SetString("234234238947234789234789234789234", 10)
	require.True(t, ok)
	input := &foo{
		A: biggun,
		B: big.NewInt(34),
	}

	bs, err := Encode(input)
	require.NoError(t, err)

	output := new(foo)
	err = Decode(bs, output)
	require.NoError(t, err)

	require.Equal(t, input, output)
}

// Order matters for serialisation
type EthRawTx struct {
	Sequence uint64   `json:"nonce"`
	GasPrice uint64   `json:"gasPrice"`
	GasLimit uint64   `json:"gasLimit"`
	To       []byte   `json:"to"`
	Amount   *big.Int `json:"value"`
	Data     []byte   `json:"data"`
	ChainID  *big.Int `json:"chainID"`

	V *big.Int
	R *big.Int
	S *big.Int
}

func TestEthRawTx(t *testing.T) {
	bigly, ok := new(big.Int).SetString("234589034578907683457689234545678235789003476899", 10)
	require.True(t, ok)

	to := crypto.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	rawTx := &EthRawTx{
		Sequence: 1,
		GasPrice: 1,
		GasLimit: 1,
		To:       to[:],
		Amount:   big.NewInt(232),
		Data:     []byte{1, 3, 4},
		ChainID:  crypto.GetEthChainID("flgoo"),
		V:        big.NewInt(272),
		R:        bigly,
		S:        bigly,
	}

	bs, err := Encode(rawTx)
	require.NoError(t, err)

	rawTxOut := new(EthRawTx)
	err = Decode(bs, rawTxOut)
	require.NoError(t, err)

	require.Equal(t, rawTx, rawTxOut)
}
