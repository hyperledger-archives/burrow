package state

// Since we have an initial save of our state forest we start one version ahead of the block height, but from then on
// we should track height by this offset.
const VersionOffset = int64(1)

func VersionAtHeight(height uint64) int64 {
	return int64(height) + VersionOffset
}

func HeightAtVersion(version int64) uint64 {
	return uint64(version - VersionOffset)
}
