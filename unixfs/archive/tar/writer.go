package tar

import (
	"archive/tar"
	"io"
	"path"
	"time"

	proto "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/github.com/gogo/protobuf/proto"
	cxt "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"

	mdag "github.com/djbarber/ipfs-hack/merkledag"
	ft "github.com/djbarber/ipfs-hack/unixfs"
	uio "github.com/djbarber/ipfs-hack/unixfs/io"
	upb "github.com/djbarber/ipfs-hack/unixfs/pb"
)

// Writer is a utility structure that helps to write
// unixfs merkledag nodes as a tar archive format.
// It wraps any io.Writer.
type Writer struct {
	Dag  mdag.DAGService
	TarW *tar.Writer

	ctx cxt.Context
}

// NewWriter wraps given io.Writer.
func NewWriter(ctx cxt.Context, dag mdag.DAGService, archive bool, compression int, w io.Writer) (*Writer, error) {
	return &Writer{
		Dag:  dag,
		TarW: tar.NewWriter(w),
		ctx:  ctx,
	}, nil
}

func (w *Writer) writeDir(nd *mdag.Node, fpath string) error {
	if err := writeDirHeader(w.TarW, fpath); err != nil {
		return err
	}

	for i, ng := range w.Dag.GetDAG(w.ctx, nd) {
		child, err := ng.Get(w.ctx)
		if err != nil {
			return err
		}

		npath := path.Join(fpath, nd.Links[i].Name)
		if err := w.WriteNode(child, npath); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) writeFile(nd *mdag.Node, pb *upb.Data, fpath string) error {
	if err := writeFileHeader(w.TarW, fpath, pb.GetFilesize()); err != nil {
		return err
	}

	dagr := uio.NewDataFileReader(w.ctx, nd, pb, w.Dag)
	if _, err := dagr.WriteTo(w.TarW); err != nil {
		return err
	}
	w.TarW.Flush()
	return nil
}

func (w *Writer) WriteNode(nd *mdag.Node, fpath string) error {
	pb := new(upb.Data)
	if err := proto.Unmarshal(nd.Data, pb); err != nil {
		return err
	}

	switch pb.GetType() {
	case upb.Data_Metadata:
		fallthrough
	case upb.Data_Directory:
		return w.writeDir(nd, fpath)
	case upb.Data_Raw:
		fallthrough
	case upb.Data_File:
		return w.writeFile(nd, pb, fpath)
	case upb.Data_Symlink:
		return writeSymlinkHeader(w.TarW, string(pb.GetData()), fpath)
	default:
		return ft.ErrUnrecognizedType
	}
}

func (w *Writer) Close() error {
	return w.TarW.Close()
}

func writeDirHeader(w *tar.Writer, fpath string) error {
	return w.WriteHeader(&tar.Header{
		Name:     fpath,
		Typeflag: tar.TypeDir,
		Mode:     0777,
		ModTime:  time.Now(),
		// TODO: set mode, dates, etc. when added to unixFS
	})
}

func writeFileHeader(w *tar.Writer, fpath string, size uint64) error {
	return w.WriteHeader(&tar.Header{
		Name:     fpath,
		Size:     int64(size),
		Typeflag: tar.TypeReg,
		Mode:     0644,
		ModTime:  time.Now(),
		// TODO: set mode, dates, etc. when added to unixFS
	})
}

func writeSymlinkHeader(w *tar.Writer, target, fpath string) error {
	return w.WriteHeader(&tar.Header{
		Name:     fpath,
		Linkname: target,
		Mode:     0777,
		Typeflag: tar.TypeSymlink,
	})
}
