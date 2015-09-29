package ipns

import (
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/djbarber/ipfs-hack/core"
	mdag "github.com/djbarber/ipfs-hack/merkledag"
	nsys "github.com/djbarber/ipfs-hack/namesys"
	ci "github.com/djbarber/ipfs-hack/p2p/crypto"
	path "github.com/djbarber/ipfs-hack/path"
	ft "github.com/djbarber/ipfs-hack/unixfs"
)

// InitializeKeyspace sets the ipns record for the given key to
// point to an empty directory.
func InitializeKeyspace(n *core.IpfsNode, key ci.PrivKey) error {
	emptyDir := &mdag.Node{Data: ft.FolderPBData()}
	nodek, err := n.DAG.Add(emptyDir)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(n.Context())
	defer cancel()

	err = n.Pinning.Pin(ctx, emptyDir, false)
	if err != nil {
		return err
	}

	err = n.Pinning.Flush()
	if err != nil {
		return err
	}

	pub := nsys.NewRoutingPublisher(n.Routing)
	if err := pub.Publish(ctx, key, path.FromKey(nodek)); err != nil {
		return err
	}

	return nil
}
