package coremock

import (
	"net"

	"github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	syncds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"

	commands "github.com/djbarber/ipfs-hack/commands"
	core "github.com/djbarber/ipfs-hack/core"
	metrics "github.com/djbarber/ipfs-hack/metrics"
	host "github.com/djbarber/ipfs-hack/p2p/host"
	mocknet "github.com/djbarber/ipfs-hack/p2p/net/mock"
	peer "github.com/djbarber/ipfs-hack/p2p/peer"
	"github.com/djbarber/ipfs-hack/repo"
	config "github.com/djbarber/ipfs-hack/repo/config"
	ds2 "github.com/djbarber/ipfs-hack/util/datastore2"
	testutil "github.com/djbarber/ipfs-hack/util/testutil"
)

// NewMockNode constructs an IpfsNode for use in tests.
func NewMockNode() (*core.IpfsNode, error) {
	ctx := context.Background()

	// effectively offline, only peer in its network
	return core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Host:   MockHostOption(mocknet.New(ctx)),
	})
}

func MockHostOption(mn mocknet.Mocknet) core.HostOption {
	return func(ctx context.Context, id peer.ID, ps peer.Peerstore, bwr metrics.Reporter, fs []*net.IPNet) (host.Host, error) {
		return mn.AddPeerWithPeerstore(id, ps)
	}
}

func MockCmdsCtx() (commands.Context, error) {
	// Generate Identity
	ident, err := testutil.RandIdentity()
	if err != nil {
		return commands.Context{}, err
	}
	p := ident.ID()

	conf := config.Config{
		Identity: config.Identity{
			PeerID: p.String(),
		},
	}

	r := &repo.Mock{
		D: ds2.CloserWrap(syncds.MutexWrap(datastore.NewMapDatastore())),
		C: conf,
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo: r,
	})

	return commands.Context{
		Online:     true,
		ConfigRoot: "/tmp/.mockipfsconfig",
		LoadConfig: func(path string) (*config.Config, error) {
			return &conf, nil
		},
		ConstructNode: func() (*core.IpfsNode, error) {
			return node, nil
		},
	}, nil
}
