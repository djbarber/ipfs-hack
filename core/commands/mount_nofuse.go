// +build linux darwin freebsd
// +build nofuse

package commands

import (
	"errors"

	cmds "github.com/djbarber/ipfs-hack/commands"
	"github.com/djbarber/ipfs-hack/core"
)

var MountCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Mounts IPFS to the filesystem (disabled)",
		ShortDescription: `
This version of ipfs is compiled without fuse support, which is required
for mounting. If you'd like to be able to mount, please use a version of
ipfs compiled with fuse.

For the latest instructions, please check the project's repository:
  http://github.com/djbarber/ipfs-hack
`,
	},
}

func Mount(node *core.IpfsNode, fsdir, nsdir string) error {
	return errors.New("not compiled in")
}
