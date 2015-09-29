package reprovide_test

import (
	"testing"

	ds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	dssync "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	blocks "github.com/djbarber/ipfs-hack/blocks"
	blockstore "github.com/djbarber/ipfs-hack/blocks/blockstore"
	mock "github.com/djbarber/ipfs-hack/routing/mock"
	testutil "github.com/djbarber/ipfs-hack/util/testutil"

	. "github.com/djbarber/ipfs-hack/exchange/reprovide"
)

func TestReprovide(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mrserv := mock.NewServer()

	idA := testutil.RandIdentityOrFatal(t)
	idB := testutil.RandIdentityOrFatal(t)

	clA := mrserv.Client(idA)
	clB := mrserv.Client(idB)

	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))

	blk := blocks.NewBlock([]byte("this is a test"))
	bstore.Put(blk)

	reprov := NewReprovider(clA, bstore)
	err := reprov.Reprovide(ctx)
	if err != nil {
		t.Fatal(err)
	}

	provs, err := clB.FindProviders(ctx, blk.Key())
	if err != nil {
		t.Fatal(err)
	}

	if len(provs) == 0 {
		t.Fatal("Should have gotten a provider")
	}

	if provs[0].ID != idA.ID() {
		t.Fatal("Somehow got the wrong peer back as a provider.")
	}
}
