package integrationtest

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"

	context "github.com/djbarber/ipfs-hack/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/djbarber/ipfs-hack/core"
	coreunix "github.com/djbarber/ipfs-hack/core/coreunix"
	mock "github.com/djbarber/ipfs-hack/core/mock"
	mocknet "github.com/djbarber/ipfs-hack/p2p/net/mock"
	"github.com/djbarber/ipfs-hack/p2p/peer"
	"github.com/djbarber/ipfs-hack/thirdparty/unit"
	testutil "github.com/djbarber/ipfs-hack/util/testutil"
)

func BenchmarkCat1MB(b *testing.B) { benchmarkVarCat(b, unit.MB*1) }
func BenchmarkCat2MB(b *testing.B) { benchmarkVarCat(b, unit.MB*2) }
func BenchmarkCat4MB(b *testing.B) { benchmarkVarCat(b, unit.MB*4) }

func benchmarkVarCat(b *testing.B, size int64) {
	data := RandomBytes(size)
	b.SetBytes(size)
	for n := 0; n < b.N; n++ {
		err := benchCat(b, data, instant)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchCat(b *testing.B, data []byte, conf testutil.LatencyConfig) error {
	b.StopTimer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create network
	mn := mocknet.New(ctx)
	mn.SetLinkDefaults(mocknet.LinkOptions{
		Latency: conf.NetworkLatency,
		// TODO add to conf. This is tricky because we want 0 values to be functional.
		Bandwidth: math.MaxInt32,
	})

	adder, err := core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Host:   mock.MockHostOption(mn),
	})
	if err != nil {
		return err
	}
	defer adder.Close()

	catter, err := core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Host:   mock.MockHostOption(mn),
	})
	if err != nil {
		return err
	}
	defer catter.Close()

	err = mn.LinkAll()
	if err != nil {
		return err
	}

	bs1 := []peer.PeerInfo{adder.Peerstore.PeerInfo(adder.Identity)}
	bs2 := []peer.PeerInfo{catter.Peerstore.PeerInfo(catter.Identity)}

	if err := catter.Bootstrap(core.BootstrapConfigWithPeers(bs1)); err != nil {
		return err
	}
	if err := adder.Bootstrap(core.BootstrapConfigWithPeers(bs2)); err != nil {
		return err
	}

	added, err := coreunix.Add(adder, bytes.NewReader(data))
	if err != nil {
		return err
	}

	b.StartTimer()
	readerCatted, err := coreunix.Cat(catter, added)
	if err != nil {
		return err
	}

	// verify
	bufout := new(bytes.Buffer)
	io.Copy(bufout, readerCatted)
	if 0 != bytes.Compare(bufout.Bytes(), data) {
		return errors.New("catted data does not match added data")
	}
	return nil
}
