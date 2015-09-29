package coreunix

import (
	"bytes"
	"io/ioutil"
	"testing"

	ds "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	dssync "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"

	bstore "github.com/djbarber/ipfs-hack/blocks/blockstore"
	key "github.com/djbarber/ipfs-hack/blocks/key"
	bserv "github.com/djbarber/ipfs-hack/blockservice"
	core "github.com/djbarber/ipfs-hack/core"
	offline "github.com/djbarber/ipfs-hack/exchange/offline"
	importer "github.com/djbarber/ipfs-hack/importer"
	chunk "github.com/djbarber/ipfs-hack/importer/chunk"
	merkledag "github.com/djbarber/ipfs-hack/merkledag"
	ft "github.com/djbarber/ipfs-hack/unixfs"
	uio "github.com/djbarber/ipfs-hack/unixfs/io"
	u "github.com/djbarber/ipfs-hack/util"
)

func getDagserv(t *testing.T) merkledag.DAGService {
	db := dssync.MutexWrap(ds.NewMapDatastore())
	bs := bstore.NewBlockstore(db)
	blockserv := bserv.New(bs, offline.Exchange(bs))
	return merkledag.NewDAGService(blockserv)
}

func TestMetadata(t *testing.T) {
	ctx := context.Background()
	// Make some random node
	ds := getDagserv(t)
	data := make([]byte, 1000)
	u.NewTimeSeededRand().Read(data)
	r := bytes.NewReader(data)
	nd, err := importer.BuildDagFromReader(ds, chunk.DefaultSplitter(r), nil)
	if err != nil {
		t.Fatal(err)
	}

	k, err := nd.Key()
	if err != nil {
		t.Fatal(err)
	}

	m := new(ft.Metadata)
	m.MimeType = "THIS IS A TEST"

	// Such effort, many compromise
	ipfsnode := &core.IpfsNode{DAG: ds}

	mdk, err := AddMetadataTo(ipfsnode, k.B58String(), m)
	if err != nil {
		t.Fatal(err)
	}

	rec, err := Metadata(ipfsnode, mdk)
	if err != nil {
		t.Fatal(err)
	}
	if rec.MimeType != m.MimeType {
		t.Fatalf("something went wrong in conversion: '%s' != '%s'", rec.MimeType, m.MimeType)
	}

	retnode, err := ds.Get(ctx, key.B58KeyDecode(mdk))
	if err != nil {
		t.Fatal(err)
	}

	ndr, err := uio.NewDagReader(ctx, retnode, ds)
	if err != nil {
		t.Fatal(err)
	}

	out, err := ioutil.ReadAll(ndr)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out, data) {
		t.Fatal("read incorrect data")
	}
}
