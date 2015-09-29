package coreunix

import (
	key "github.com/djbarber/ipfs-hack/blocks/key"
	core "github.com/djbarber/ipfs-hack/core"
	dag "github.com/djbarber/ipfs-hack/merkledag"
	ft "github.com/djbarber/ipfs-hack/unixfs"
)

func AddMetadataTo(n *core.IpfsNode, skey string, m *ft.Metadata) (string, error) {
	ukey := key.B58KeyDecode(skey)

	nd, err := n.DAG.Get(n.Context(), ukey)
	if err != nil {
		return "", err
	}

	mdnode := new(dag.Node)
	mdata, err := ft.BytesForMetadata(m)
	if err != nil {
		return "", err
	}

	mdnode.Data = mdata
	if err := mdnode.AddNodeLinkClean("file", nd); err != nil {
		return "", err
	}

	nk, err := n.DAG.Add(mdnode)
	if err != nil {
		return "", err
	}

	return nk.B58String(), nil
}

func Metadata(n *core.IpfsNode, skey string) (*ft.Metadata, error) {
	ukey := key.B58KeyDecode(skey)

	nd, err := n.DAG.Get(n.Context(), ukey)
	if err != nil {
		return nil, err
	}

	return ft.MetadataFromBytes(nd.Data)
}
