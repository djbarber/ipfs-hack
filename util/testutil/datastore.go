package testutil

import (
	"github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	syncds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	ds2 "github.com/djbarber/ipfs-hack/util/datastore2"
)

func ThreadSafeCloserMapDatastore() ds2.ThreadSafeDatastoreCloser {
	return ds2.CloserWrap(syncds.MutexWrap(datastore.NewMapDatastore()))
}
