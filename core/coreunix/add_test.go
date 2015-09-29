package coreunix

import (
	"os"
	"path"
	"testing"

	"github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/djbarber/ipfs-hack/core"
	"github.com/djbarber/ipfs-hack/repo"
	"github.com/djbarber/ipfs-hack/repo/config"
	"github.com/djbarber/ipfs-hack/util/testutil"
)

func TestAddRecursive(t *testing.T) {
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	r := &repo.Mock{
		C: config.Config{
			Identity: config.Identity{
				PeerID: "Qmfoo", // required by offline node
			},
		},
		D: testutil.ThreadSafeCloserMapDatastore(),
	}
	node, err := core.NewNode(context.Background(), &core.BuildCfg{Repo: r})
	if err != nil {
		t.Fatal(err)
	}
	if k, err := AddR(node, path.Join(here, "test_data")); err != nil {
		t.Fatal(err)
	} else if k != "QmWCCga8AbTyfAQ7pTnGT6JgmRMAB3Qp8ZmTEFi5q5o8jC" {
		t.Fatal("keys do not match")
	}
}
