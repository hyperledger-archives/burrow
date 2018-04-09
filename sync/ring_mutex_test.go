package sync

import (
	"encoding/base64"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingMutexXXHash_Lock(t *testing.T) {
	mutexCount := 10
	numAddresses := byte(20)
	mtxs := []*RingMutex{NewRingMutexXXHash(mutexCount)}

	for _, mtx := range mtxs {
		// Using fewer mutexes than addresses to lock against should cause contention
		writeCh := make(chan []byte)
		checksum := 0

		// We'll try to acquire a locks on all of our unique addresses, knowing that
		// some of them will share an underlying RWMutex
		for i := byte(0); i < numAddresses; i++ {
			address := []byte{i}
			go func() {
				mtx.Lock(address)
				writeCh <- address
			}()
		}

		// We should receive a message from all of those addresses for which we could
		// acquire a lock, this should be almost surely deterministic since we are
		// launching our goroutines sequentially from a single goroutine (if this bit
		// breaks we can add a short pause between the 'go' statements above, for the
		// purposes of the predictability of this test)
		addresses := receiveAddresses(writeCh)
		checksum += len(addresses)
		// we hit lock contention on the tenth address so get 9 back
		assert.Equal(t, 9, len(addresses))
		// Unlock the 9 locked mutexes
		unlockAddresses(mtx, addresses)

		// Which should trigger another batch to make progress
		addresses = receiveAddresses(writeCh)
		checksum += len(addresses)
		// Again the number we get back (but not the order) should be deterministic
		// because we are unlocking sequentially from a single goroutine
		assert.Equal(t, 7, len(addresses))
		unlockAddresses(mtx, addresses)

		// And again
		addresses = receiveAddresses(writeCh)
		checksum += len(addresses)
		assert.Equal(t, 3, len(addresses))
		unlockAddresses(mtx, addresses)

		// And so on
		addresses = receiveAddresses(writeCh)
		checksum += len(addresses)
		assert.Equal(t, 1, len(addresses))
		unlockAddresses(mtx, addresses)

		// Until we have unblocked all of the goroutines we released
		addresses = receiveAddresses(writeCh)
		checksum += len(addresses)
		assert.Equal(t, 0, len(addresses))
		unlockAddresses(mtx, addresses)
		checksum += len(addresses)

		// Check we've heard back from all of them
		assert.EqualValues(t, numAddresses, checksum)
	}
}

func TestRingMutex_XXHash(t *testing.T) {
	mtx := NewRingMutexXXHash(10)
	address, err := base64.StdEncoding.DecodeString("/+ulTkCzpYg2ePaZtqS8dycJBLY9387yZPst8LX5YL0=")
	assert.NoError(t, err)
	assert.EqualValues(t, 8509033946529530334, mtx.hash(address))
}

func receiveAddresses(returnCh chan []byte) [][]byte {
	var addresses [][]byte
	for {
		select {
		case address := <-returnCh:
			addresses = append(addresses, address)
		case <-time.After(50 * time.Millisecond):
			return addresses
		}
	}
}
func unlockAddresses(mtx *RingMutex, addresses [][]byte) {
	for _, address := range addresses {
		mtx.Unlock(address)
	}
}
