package blockchain

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/stretchr/testify/assert"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tm_types "github.com/tendermint/tendermint/types"
)

func TestValidatorSet(t *testing.T) {
	publicKeys := generatePublickKeys()
	validators := make([]acm.Validator, 6)
	validators[0] = acm.NewValidator(publicKeys[0], 100, 1)
	validators[1] = acm.NewValidator(publicKeys[1], 200, 1)
	validators[2] = acm.NewValidator(publicKeys[2], 300, 1)
	validators[3] = acm.NewValidator(publicKeys[3], 400, 1)
	validators[4] = acm.NewValidator(publicKeys[4], 500, 1)
	validators[5] = acm.NewValidator(publicKeys[5], 600, 1)

	vs := newValidatorSet(8, validators)

	val := acm.NewValidator(publickKeyFromSecrt("z"), 100, 1)

	err := vs.LeaveFromTheSet(val)
	assert.Error(t, err)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, false, vs.IsValidatorInSet(val.Address()))
	err = vs.JoinToTheSet(val)
	assert.NoError(t, err)
	assert.Equal(t, 7, vs.TotalPower())
	assert.Equal(t, true, vs.IsValidatorInSet(val.Address()))
	err = vs.JoinToTheSet(val)
	assert.Error(t, err)
	vs.LeaveFromTheSet(val)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, false, vs.IsValidatorInSet(val.Address()))
}

type _validatorListProxyMock struct {
	height        int64
	validatorSets [][]*tm_types.Validator
}

func newValidatorListProxyMock() *_validatorListProxyMock {

	vp := &_validatorListProxyMock{}
	publicKeys := generatePublickKeys()
	validators := make([]*tm_types.Validator, len(publicKeys))

	for i, p := range publicKeys {
		validators[i] = tm_types.NewValidator(p.PubKey, 1)
	}

	/// round:1, power:4
	/// <- validator[0,1,2,3] joined
	vp.NextRound(validators[0:4])

	/// round:2, power:5
	/// <- validator[4] joined
	vp.NextRound(validators[0:5])

	/// round:3, power:6
	/// <- validator[5] joined
	vp.NextRound(validators[0:6])

	/// round:4, power:7
	/// <- validator[6] joined
	vp.NextRound(validators[0:7])

	/// round:5, power:8
	/// <- validator[7] joined
	vp.NextRound(validators[0:8])

	/// round:6, power:8 (no change)
	vp.NextRound(validators[0:8])

	/// round:7
	/// -> validator[0] left
	/// <- validator[8] joined
	vp.NextRound(validators[1:9])

	/// round:8
	/// -> validator[1] left
	/// <- validator[9,10,11,12] joined
	vp.NextRound(validators[2:13])

	/// round:9
	/// -> validator[2] left
	/// <- validator[13] joined
	vp.NextRound(validators[3:14])

	/// round:10
	/// -> validator[3] left
	vp.NextRound(validators[4:14])

	/// round:11
	/// -> validator[4] left
	vp.NextRound(validators[5:14])

	/// round:12
	/// -> validator[5] left
	vp.NextRound(validators[6:14])

	/// round:13
	/// -> validator[6] left
	vp.NextRound(validators[6:14])

	/// round:14
	vp.NextRound(validators[6:14])

	return vp
}

func (vlp _validatorListProxyMock) Validators(height int64) (*ctypes.ResultValidators, error) {
	var result ctypes.ResultValidators
	result.Validators = vlp.validatorSets[height-1]
	result.BlockHeight = height

	return &result, nil
}

func (vlp _validatorListProxyMock) Validators2(height int64) []*tm_types.Validator {
	result, _ := vlp.Validators(height)

	return result.Validators
}

func (vlp *_validatorListProxyMock) NextRound(validators []*tm_types.Validator) {
	validators2 := make([]*tm_types.Validator, len(validators))
	copy(validators2, validators)

	vlp.height++
	vlp.validatorSets = append(vlp.validatorSets, validators2)
}

func TestAdjusting(t *testing.T) {

	vp := newValidatorListProxyMock()
	publicKeys := generatePublickKeys()
	validators := make([]acm.Validator, len(publicKeys))
	var err error

	for i, p := range publicKeys {
		validators[i] = acm.NewValidator(p, 1, 100)
	}

	vs := validatorSet{maximumPower: 8, validators: validators[0:4], proxy: vp}

	// -----------------------------------------
	vs.JoinToTheSet(validators[4])
	err = vs.AdjustPower(2)

	assert.NoError(t, err)
	assert.Equal(t, 5, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(2)))

	// println(fmt.Sprintf("%v", vs.Validators()))
	// println(fmt.Sprintf("%v", vp.Validators2))

	// -----------------------------------------
	vs.JoinToTheSet(validators[5])
	err = vs.AdjustPower(3)

	assert.NoError(t, err)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(3)))

	// -----------------------------------------
	vs.JoinToTheSet(validators[6])
	err = vs.AdjustPower(4)

	assert.NoError(t, err)
	assert.Equal(t, 7, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(4)))

	// -----------------------------------------
	vs.JoinToTheSet(validators[7])
	err = vs.AdjustPower(5)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(5)))

	// -----------------------------------------
	err = vs.AdjustPower(6)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(6)))

	// -----------------------------------------
	vs.JoinToTheSet(validators[8])
	err = vs.AdjustPower(7)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(7)))

	// -----------------------------------------
	vs.JoinToTheSet(validators[9])
	vs.JoinToTheSet(validators[10])
	vs.JoinToTheSet(validators[11])
	vs.JoinToTheSet(validators[12])
	err = vs.AdjustPower(8)

	assert.NoError(t, err)
	assert.Equal(t, 11, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(8)))

	// -----------------------------------------
	vs.JoinToTheSet(validators[9])
	vs.JoinToTheSet(validators[10])
	vs.JoinToTheSet(validators[11])
	vs.JoinToTheSet(validators[12])
	vs.JoinToTheSet(validators[13])
	err = vs.AdjustPower(9)

	assert.NoError(t, err)
	assert.Equal(t, 11, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(9)))

	// -----------------------------------------
	err = vs.AdjustPower(10)

	assert.NoError(t, err)
	assert.Equal(t, 10, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(10)))

	// -----------------------------------------
	err = vs.AdjustPower(11)

	assert.NoError(t, err)
	assert.Equal(t, 9, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(11)))

	// -----------------------------------------
	err = vs.AdjustPower(12)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(12)))

	// -----------------------------------------
	err = vs.AdjustPower(13)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(13)))

	// -----------------------------------------
	err = vs.AdjustPower(14)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), vp.Validators2(14)))

}

func compareValidators(validators1 []acm.Validator, validators2 []*tm_types.Validator) bool {

	if len(validators1) != len(validators2) {
		return false
	}

	for _, v1 := range validators1 {
		found := false
		for _, v2 := range validators1 {
			if v1.Address() == v2.Address() {
				found = true
				break
			}
		}
		if found == false {
			return false
		}
	}

	return true
}

func publickKeyFromSecrt(secret string) acm.PublicKey {
	return acm.GeneratePrivateAccountFromSecret(secret).PublicKey()
}

func generatePublickKeys() []acm.PublicKey {
	publicKey := make([]acm.PublicKey, 26)

	/// sorted by address
	publicKey[0] = publickKeyFromSecrt("m")  //  18A71D0D81CEEBF548019C4BC24BB6F5B4E1361F
	publicKey[1] = publickKeyFromSecrt("w")  //  1B4557CC1850966A88DCF7094F18ACC6756F1250
	publicKey[2] = publickKeyFromSecrt("c")  //  366FC725E46FFDE3E63152AEA34B6EA15816D47D
	publicKey[3] = publickKeyFromSecrt("x")  //  3F56ED107D8A808AEB2AAB523B72F7C37C812894
	publicKey[4] = publickKeyFromSecrt("v")  //  4203FC4AE98849F0A1B6CB7E027FDE2FABD7AC62
	publicKey[5] = publickKeyFromSecrt("a")  //  433CA69C9F597C9CD105740B04FC8CBFF206B587
	publicKey[6] = publickKeyFromSecrt("r")  //  4861B368170E44623B86359B234CD4C485205678
	publicKey[7] = publickKeyFromSecrt("z")  //  494F9624293B91E23C3D2AD946BB020F79D73CA8
	publicKey[8] = publickKeyFromSecrt("t")  //  61BE4158B77C63BF69C8AF6733614F67C7DB45BA
	publicKey[9] = publickKeyFromSecrt("n")  //  644CD981E309F6230F71C6434164205C64F82463
	publicKey[10] = publickKeyFromSecrt("k") //  695AB9E2D56F83EA8403C007F17CCCB37A398594
	publicKey[11] = publickKeyFromSecrt("i") //  7A37293D9152D3BE4A61DC4E79E7357421F212CC
	publicKey[12] = publickKeyFromSecrt("j") //  8426FF76304CEAF18EE662B75E1B277303CD498C
	publicKey[13] = publickKeyFromSecrt("d") //  8DFB3FDB0F0852D11BD58C231F09CEE35B78A376
	publicKey[14] = publickKeyFromSecrt("q") //  91B3B57CA5921AC9F31A359F709A517F1D37A709
	publicKey[15] = publickKeyFromSecrt("h") //  9CB5809A3FC3E9201C9039123F4BA39BD87F76FD
	publicKey[16] = publickKeyFromSecrt("b") //  9E9AC3380A4941075BD2DDD534D7524BDDA6BB15
	publicKey[17] = publickKeyFromSecrt("s") //  A49B77DB290F20764C70B46C16FB5D6801F70362
	publicKey[18] = publickKeyFromSecrt("u") //  A83D2DABD3477EB60DA9A25343173BE7B4454728
	publicKey[19] = publickKeyFromSecrt("y") //  AA6357CBBA4CF942178D03A02AC258BE168B7BCE
	publicKey[20] = publickKeyFromSecrt("p") //  B044193ACBE2144DE94A7DA85E6A86DFF359C85E
	publicKey[21] = publickKeyFromSecrt("l") //  C1461B8B1DC8D838A44B44E6ED423708258508DB
	publicKey[22] = publickKeyFromSecrt("e") //  C75F161831E073F28FDA2C8C6DE20DFBFE277CA1
	publicKey[23] = publickKeyFromSecrt("f") //  CC7B4572884B5C99D546FB8170A2A650DDCBA78E
	publicKey[24] = publickKeyFromSecrt("o") //  E68964E9A5DFCE79A04D7E1B5AEEC5795C94BC73
	publicKey[25] = publickKeyFromSecrt("g") //  FA987DA7B094D32392AF377A5079FC1D30DCC214

	return publicKey
}
