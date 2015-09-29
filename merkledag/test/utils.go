package mdutils

import (
	ds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	dssync "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	"github.com/djbarber/ipfs-hack/blocks/blockstore"
	bsrv "github.com/djbarber/ipfs-hack/blockservice"
	"github.com/djbarber/ipfs-hack/exchange/offline"
	dag "github.com/djbarber/ipfs-hack/merkledag"
)

func Mock() dag.DAGService {
	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))
	bserv := bsrv.New(bstore, offline.Exchange(bstore))
	return dag.NewDAGService(bserv)
}
