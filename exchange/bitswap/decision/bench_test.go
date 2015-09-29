package decision

import (
	"math"
	"testing"

	key "github.com/djbarber/ipfs-hack/blocks/key"
	"github.com/djbarber/ipfs-hack/exchange/bitswap/wantlist"
	"github.com/djbarber/ipfs-hack/p2p/peer"
	"github.com/djbarber/ipfs-hack/util/testutil"
)

// FWIW: At the time of this commit, including a timestamp in task increases
// time cost of Push by 3%.
func BenchmarkTaskQueuePush(b *testing.B) {
	q := newPRQ()
	peers := []peer.ID{
		testutil.RandPeerIDFatal(b),
		testutil.RandPeerIDFatal(b),
		testutil.RandPeerIDFatal(b),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Push(wantlist.Entry{Key: key.Key(i), Priority: math.MaxInt32}, peers[i%len(peers)])
	}
}
