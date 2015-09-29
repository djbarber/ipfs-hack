package testutil

import (
	"testing"

	metrics "github.com/djbarber/ipfs-hack/metrics"
	bhost "github.com/djbarber/ipfs-hack/p2p/host/basic"
	inet "github.com/djbarber/ipfs-hack/p2p/net"
	swarm "github.com/djbarber/ipfs-hack/p2p/net/swarm"
	peer "github.com/djbarber/ipfs-hack/p2p/peer"
	tu "github.com/djbarber/ipfs-hack/util/testutil"

	ma "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
)

func GenSwarmNetwork(t *testing.T, ctx context.Context) *swarm.Network {
	p := tu.RandPeerNetParamsOrFatal(t)
	ps := peer.NewPeerstore()
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	n, err := swarm.NewNetwork(ctx, []ma.Multiaddr{p.Addr}, p.ID, ps, metrics.NewBandwidthCounter())
	if err != nil {
		t.Fatal(err)
	}
	ps.AddAddrs(p.ID, n.ListenAddresses(), peer.PermanentAddrTTL)
	return n
}

func DivulgeAddresses(a, b inet.Network) {
	id := a.LocalPeer()
	addrs := a.Peerstore().Addrs(id)
	b.Peerstore().AddAddrs(id, addrs, peer.PermanentAddrTTL)
}

func GenHostSwarm(t *testing.T, ctx context.Context) *bhost.BasicHost {
	n := GenSwarmNetwork(t, ctx)
	return bhost.New(n)
}

var RandPeerID = tu.RandPeerID
