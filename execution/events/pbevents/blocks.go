package pbevents

import "github.com/hyperledger/burrow/execution/events"

// Get bounds suitable for events.Provider
func (br *BlockRange) Bounds(latestBlockHeight uint64) (startKey, endKey events.Key, streaming bool) {
	// End bound is exclusive in state.GetEvents so we increment the height
	return br.GetStart().Key(latestBlockHeight), br.GetEnd().Key(latestBlockHeight).IncHeight(),
		br.GetEnd().GetType() == Bound_STREAM
}

func (b *Bound) Key(latestBlockHeight uint64) events.Key {
	return events.NewKey(b.Bound(latestBlockHeight), 0)
}

func (b *Bound) Bound(latestBlockHeight uint64) uint64 {
	if b == nil {
		return latestBlockHeight
	}
	switch b.Type {
	case Bound_ABSOLUTE:
		return b.GetIndex()
	case Bound_RELATIVE:
		if b.Index < latestBlockHeight {
			return latestBlockHeight - b.Index
		}
		return 0
	case Bound_FIRST:
		return 0
	case Bound_LATEST, Bound_STREAM:
		return latestBlockHeight
	default:
		return latestBlockHeight
	}
}

func AbsoluteBound(index uint64) *Bound {
	return &Bound{
		Index: index,
		Type:  Bound_ABSOLUTE,
	}
}

func RelativeBound(index uint64) *Bound {
	return &Bound{
		Index: index,
		Type:  Bound_RELATIVE,
	}
}

func LatestBound() *Bound {
	return &Bound{
		Type: Bound_LATEST,
	}
}

func StreamBound() *Bound {
	return &Bound{
		Type: Bound_STREAM,
	}
}

func NewBlockRange(start, end *Bound) *BlockRange {
	return &BlockRange{
		Start: start,
		End:   end,
	}
}
