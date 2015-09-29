package bitswap

import (
	ds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	bsnet "github.com/djbarber/ipfs-hack/exchange/bitswap/network"
	mockpeernet "github.com/djbarber/ipfs-hack/p2p/net/mock"
	peer "github.com/djbarber/ipfs-hack/p2p/peer"
	mockrouting "github.com/djbarber/ipfs-hack/routing/mock"
	testutil "github.com/djbarber/ipfs-hack/util/testutil"
)

type peernet struct {
	mockpeernet.Mocknet
	routingserver mockrouting.Server
}

func StreamNet(ctx context.Context, net mockpeernet.Mocknet, rs mockrouting.Server) (Network, error) {
	return &peernet{net, rs}, nil
}

func (pn *peernet) Adapter(p testutil.Identity) bsnet.BitSwapNetwork {
	client, err := pn.Mocknet.AddPeer(p.PrivateKey(), p.Address())
	if err != nil {
		panic(err.Error())
	}
	routing := pn.routingserver.ClientWithDatastore(context.TODO(), p, ds.NewMapDatastore())
	return bsnet.NewFromIpfsHost(client, routing)
}

func (pn *peernet) HasPeer(p peer.ID) bool {
	for _, member := range pn.Mocknet.Peers() {
		if p == member {
			return true
		}
	}
	return false
}

var _ Network = &peernet{}
