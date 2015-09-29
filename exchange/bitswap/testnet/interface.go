package bitswap

import (
	bsnet "github.com/djbarber/ipfs-hack/exchange/bitswap/network"
	peer "github.com/djbarber/ipfs-hack/p2p/peer"
	"github.com/djbarber/ipfs-hack/util/testutil"
)

type Network interface {
	Adapter(testutil.Identity) bsnet.BitSwapNetwork

	HasPeer(peer.ID) bool
}
