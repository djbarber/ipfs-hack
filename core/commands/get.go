package commands

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	gopath "path"
	"strings"

	"github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/cheggaaa/pb"

	cmds "github.com/djbarber/ipfs-hack/commands"
	core "github.com/djbarber/ipfs-hack/core"
	path "github.com/djbarber/ipfs-hack/path"
	tar "github.com/djbarber/ipfs-hack/thirdparty/tar"
	uarchive "github.com/djbarber/ipfs-hack/unixfs/archive"
)

var ErrInvalidCompressionLevel = errors.New("Compression level must be between 1 and 9")

var GetCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Download IPFS objects",
		ShortDescription: `
Retrieves the object named by <ipfs-or-ipns-path> and stores the data to disk.

By default, the output will be stored at ./<ipfs-path>, but an alternate path
can be specified with '--output=<path>' or '-o=<path>'.

To output a TAR archive instead of unpacked files, use '--archive' or '-a'.

To compress the output with GZIP compression, use '--compress' or '-C'. You
may also specify the level of compression by specifying '-l=<1-9>'.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, false, "The path to the IPFS object(s) to be outputted").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("output", "o", "The path where output should be stored"),
		cmds.BoolOption("archive", "a", "Output a TAR archive"),
		cmds.BoolOption("compress", "C", "Compress the output with GZIP compression"),
		cmds.IntOption("compression-level", "l", "The level of compression (1-9)"),
	},
	PreRun: func(req cmds.Request) error {
		_, err := getCompressOptions(req)
		return err
	},
	Run: func(req cmds.Request, res cmds.Response) {
		cmplvl, err := getCompressOptions(req)
		if err != nil {
			res.SetError(err, cmds.ErrClient)
			return
		}

		node, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
		p := path.Path(req.Arguments()[0])
		ctx := req.Context()
		dn, err := core.Resolve(ctx, node, p)
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		archive, _, _ := req.Option("archive").Bool()
		reader, err := uarchive.DagArchive(ctx, dn, p.String(), node.DAG, archive, cmplvl)
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
		res.SetOutput(reader)
	},
	PostRun: func(req cmds.Request, res cmds.Response) {
		if res.Output() == nil {
			return
		}
		outReader := res.Output().(io.Reader)
		res.SetOutput(nil)

		outPath, _, _ := req.Option("output").String()
		if len(outPath) == 0 {
			_, outPath = gopath.Split(req.Arguments()[0])
			outPath = gopath.Clean(outPath)
		}

		cmplvl, err := getCompressOptions(req)
		if err != nil {
			res.SetError(err, cmds.ErrClient)
			return
		}

		archive, _, _ := req.Option("archive").Bool()

		gw := getWriter{
			Out:         os.Stdout,
			Err:         os.Stderr,
			Archive:     archive,
			Compression: cmplvl,
		}

		if err := gw.Write(outReader, outPath); err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
	},
}

func progressBarForReader(out io.Writer, r io.Reader) (*pb.ProgressBar, *pb.Reader) {
	// setup bar reader
	// TODO: get total length of files
	bar := pb.New(0).SetUnits(pb.U_BYTES)
	bar.Output = out
	barR := bar.NewProxyReader(r)
	return bar, barR
}

type getWriter struct {
	Out io.Writer // for output to user
	Err io.Writer // for progress bar output

	Archive     bool
	Compression int
}

func (gw *getWriter) Write(r io.Reader, fpath string) error {
	if gw.Archive || gw.Compression != gzip.NoCompression {
		return gw.writeArchive(r, fpath)
	}
	return gw.writeExtracted(r, fpath)
}

func (gw *getWriter) writeArchive(r io.Reader, fpath string) error {
	// adjust file name if tar
	if gw.Archive {
		if !strings.HasSuffix(fpath, ".tar") && !strings.HasSuffix(fpath, ".tar.gz") {
			fpath += ".tar"
		}
	}

	// adjust file name if gz
	if gw.Compression != gzip.NoCompression {
		if !strings.HasSuffix(fpath, ".gz") {
			fpath += ".gz"
		}
	}

	// create file
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(gw.Out, "Saving archive to %s\n", fpath)
	bar, barR := progressBarForReader(gw.Err, r)
	bar.Start()
	defer bar.Finish()

	_, err = io.Copy(file, barR)
	return err
}

func (gw *getWriter) writeExtracted(r io.Reader, fpath string) error {
	fmt.Fprintf(gw.Out, "Saving file(s) to %s\n", fpath)
	bar, barR := progressBarForReader(gw.Err, r)
	bar.Start()
	defer bar.Finish()

	extractor := &tar.Extractor{fpath}
	return extractor.Extract(barR)
}

func getCompressOptions(req cmds.Request) (int, error) {
	cmprs, _, _ := req.Option("compress").Bool()
	cmplvl, cmplvlFound, _ := req.Option("compression-level").Int()
	switch {
	case !cmprs:
		return gzip.NoCompression, nil
	case cmprs && !cmplvlFound:
		return gzip.DefaultCompression, nil
	case cmprs && cmplvlFound && (cmplvl < 1 || cmplvl > 9):
		return gzip.NoCompression, ErrInvalidCompressionLevel
	}
	return cmplvl, nil
}
