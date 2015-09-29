package coreunix

import (
	core "github.com/djbarber/ipfs-hack/core"
	path "github.com/djbarber/ipfs-hack/path"
	uio "github.com/djbarber/ipfs-hack/unixfs/io"
)

func Cat(n *core.IpfsNode, pstr string) (*uio.DagReader, error) {
	p := path.FromString(pstr)
	dagNode, err := n.Resolver.ResolvePath(n.Context(), p)
	if err != nil {
		return nil, err
	}
	return uio.NewDagReader(n.Context(), dagNode, n.DAG)
}
