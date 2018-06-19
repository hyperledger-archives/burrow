package pbevents

func (br *BlockRange) Bounds() (start uint64, end uint64) {
	return br.GetStart().GetIndex(), br.GetEnd().GetIndex()
}

func SimpleBlockRange(start, end uint64) *BlockRange {
	return &BlockRange{
		Start: &Bound{
			Type:  Bound_ABSOLUTE,
			Index: start,
		},
		End: &Bound{
			Type:  Bound_ABSOLUTE,
			Index: end,
		},
	}
}
