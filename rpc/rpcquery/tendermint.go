package rpcquery

import (
	"encoding/json"

	"github.com/tendermint/tendermint/types"
)

type Header struct {
	Header types.Header
}

func (head *Header) MarshalJSON() ([]byte, error) {
	return json.Marshal(head)
}

func (head *Header) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, head)
}

func (head *Header) Marshal() ([]byte, error) {
	if head == nil {
		return nil, nil
	}
	return head.MarshalJSON()
}

func (head *Header) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return head.UnmarshalJSON(data)
}

func (head *Header) MarshalTo(data []byte) (int, error) {
	bs, err := head.Marshal()
	if err != nil {
		return 0, err
	}
	return copy(data, bs), nil
}

func (head *Header) Size() int {
	bs, _ := head.Marshal()
	return len(bs)
}

type Commit struct {
	Commit *types.Commit
}

func (cmm *Commit) MarshalJSON() ([]byte, error) {
	return json.Marshal(cmm)
}

func (cmm *Commit) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, cmm)
}

func (cmm *Commit) Marshal() ([]byte, error) {
	if cmm == nil {
		return nil, nil
	}
	return cmm.MarshalJSON()
}

func (cmm *Commit) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return cmm.UnmarshalJSON(data)
}

func (cmm *Commit) MarshalTo(data []byte) (int, error) {
	bs, err := cmm.Marshal()
	if err != nil {
		return 0, err
	}
	return copy(data, bs), nil
}

func (cmm *Commit) Size() int {
	bs, _ := cmm.Marshal()
	return len(bs)
}

type Validators struct {
	ValidatorSet *types.ValidatorSet
}

func (val *Validators) MarshalJSON() ([]byte, error) {
	return json.Marshal(val)
}

func (val *Validators) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, val)
}

func (val *Validators) Marshal() ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	return val.MarshalJSON()
}

func (val *Validators) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return val.UnmarshalJSON(data)
}

func (val *Validators) MarshalTo(data []byte) (int, error) {
	bs, err := val.Marshal()
	if err != nil {
		return 0, err
	}
	return copy(data, bs), nil
}

func (val *Validators) Size() int {
	bs, _ := val.Marshal()
	return len(bs)
}
