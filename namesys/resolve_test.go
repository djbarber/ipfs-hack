package namesys

import (
	"testing"

	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	key "github.com/djbarber/ipfs-hack/blocks/key"
	path "github.com/djbarber/ipfs-hack/path"
	mockrouting "github.com/djbarber/ipfs-hack/routing/mock"
	u "github.com/djbarber/ipfs-hack/util"
	testutil "github.com/djbarber/ipfs-hack/util/testutil"
)

func TestRoutingResolve(t *testing.T) {
	d := mockrouting.NewServer().Client(testutil.RandIdentityOrFatal(t))

	resolver := NewRoutingResolver(d)
	publisher := NewRoutingPublisher(d)

	privk, pubk, err := testutil.RandTestKeyPair(512)
	if err != nil {
		t.Fatal(err)
	}

	h := path.FromString("/ipfs/QmZULkCELmmk5XNfCgTnCyFgAVxBRBXyDHGGMVoLFLiXEN")
	err = publisher.Publish(context.Background(), privk, h)
	if err != nil {
		t.Fatal(err)
	}

	pubkb, err := pubk.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	pkhash := u.Hash(pubkb)
	res, err := resolver.Resolve(context.Background(), key.Key(pkhash).Pretty())
	if err != nil {
		t.Fatal(err)
	}

	if res != h {
		t.Fatal("Got back incorrect value.")
	}
}
