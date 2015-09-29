package core_test

import (
	"testing"

	core "github.com/djbarber/ipfs-hack/core"
	coremock "github.com/djbarber/ipfs-hack/core/mock"
	path "github.com/djbarber/ipfs-hack/path"
)

func TestResolveNoComponents(t *testing.T) {
	n, err := coremock.NewMockNode()
	if n == nil || err != nil {
		t.Fatal("Should have constructed a mock node", err)
	}

	_, err = core.Resolve(n.Context(), n, path.Path("/ipns/"))
	if err != path.ErrNoComponents {
		t.Fatal("Should error with no components (/ipns/).", err)
	}

	_, err = core.Resolve(n.Context(), n, path.Path("/ipfs/"))
	if err != path.ErrNoComponents {
		t.Fatal("Should error with no components (/ipfs/).", err)
	}

	_, err = core.Resolve(n.Context(), n, path.Path("/../.."))
	if err != path.ErrBadPath {
		t.Fatal("Should error with invalid path.", err)
	}
}
