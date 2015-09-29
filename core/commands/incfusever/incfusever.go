// +build !nofuse

package incfusever

import (
	fuseversion "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/jbenet/go-fuse-version"
)

var _ = fuseversion.LocalFuseSystems
