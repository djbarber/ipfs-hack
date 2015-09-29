// package importer implements utilities used to create ipfs DAGs from files
// and readers
package importer

import (
	"fmt"
	"os"

	"github.com/djbarber/ipfs-hack/commands/files"
	bal "github.com/djbarber/ipfs-hack/importer/balanced"
	"github.com/djbarber/ipfs-hack/importer/chunk"
	h "github.com/djbarber/ipfs-hack/importer/helpers"
	trickle "github.com/djbarber/ipfs-hack/importer/trickle"
	dag "github.com/djbarber/ipfs-hack/merkledag"
	"github.com/djbarber/ipfs-hack/pin"
	logging "github.com/djbarber/ipfs-hack/vendor/go-log-v1.0.0"
)

var log = logging.Logger("importer")

// Builds a DAG from the given file, writing created blocks to disk as they are
// created
func BuildDagFromFile(fpath string, ds dag.DAGService, mp pin.ManualPinner) (*dag.Node, error) {
	stat, err := os.Lstat(fpath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("`%s` is a directory", fpath)
	}

	f, err := files.NewSerialFile(fpath, fpath, stat)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return BuildDagFromReader(ds, chunk.NewSizeSplitter(f, chunk.DefaultBlockSize), BasicPinnerCB(mp))
}

func BuildDagFromReader(ds dag.DAGService, spl chunk.Splitter, ncb h.NodeCB) (*dag.Node, error) {
	// Start the splitter
	blkch, errch := chunk.Chan(spl)

	dbp := h.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: h.DefaultLinksPerBlock,
		NodeCB:   ncb,
	}

	return bal.BalancedLayout(dbp.New(blkch, errch))
}

func BuildTrickleDagFromReader(ds dag.DAGService, spl chunk.Splitter, ncb h.NodeCB) (*dag.Node, error) {
	// Start the splitter
	blkch, errch := chunk.Chan(spl)

	dbp := h.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: h.DefaultLinksPerBlock,
		NodeCB:   ncb,
	}

	return trickle.TrickleLayout(dbp.New(blkch, errch))
}

func BasicPinnerCB(p pin.ManualPinner) h.NodeCB {
	return func(n *dag.Node, last bool) error {
		k, err := n.Key()
		if err != nil {
			return err
		}

		if last {
			p.PinWithMode(k, pin.Recursive)
			return p.Flush()
		} else {
			p.PinWithMode(k, pin.Indirect)
			return nil
		}
	}
}

func PinIndirectCB(p pin.ManualPinner) h.NodeCB {
	return func(n *dag.Node, last bool) error {
		k, err := n.Key()
		if err != nil {
			return err
		}

		p.PinWithMode(k, pin.Indirect)
		return nil
	}
}
