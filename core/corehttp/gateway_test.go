package corehttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"
	core "github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	namesys "github.com/ipfs/go-ipfs/namesys"
	ci "github.com/ipfs/go-ipfs/p2p/crypto"
	path "github.com/ipfs/go-ipfs/path"
	repo "github.com/ipfs/go-ipfs/repo"
	config "github.com/ipfs/go-ipfs/repo/config"
	testutil "github.com/ipfs/go-ipfs/util/testutil"
)

type mockNamesys map[string]path.Path

func (m mockNamesys) Resolve(ctx context.Context, name string) (value path.Path, err error) {
	return m.ResolveN(ctx, name, namesys.DefaultDepthLimit)
}

func (m mockNamesys) ResolveN(ctx context.Context, name string, depth int) (value path.Path, err error) {
	p, ok := m[name]
	if !ok {
		return "", namesys.ErrResolveFailed
	}
	return p, nil
}

func (m mockNamesys) Publish(ctx context.Context, name ci.PrivKey, value path.Path) error {
	return errors.New("not implemented for mockNamesys")
}

func newNodeWithMockNamesys(ns mockNamesys) (*core.IpfsNode, error) {
	c := config.Config{
		Identity: config.Identity{
			PeerID: "Qmfoo", // required by offline node
		},
	}
	r := &repo.Mock{
		C: c,
		D: testutil.ThreadSafeCloserMapDatastore(),
	}
	n, err := core.NewNode(context.Background(), &core.BuildCfg{Repo: r})
	if err != nil {
		return nil, err
	}
	n.Namesys = ns
	return n, nil
}

type delegatedHandler struct {
	http.Handler
}

func (dh *delegatedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dh.Handler.ServeHTTP(w, r)
}

func doWithoutRedirect(req *http.Request) (*http.Response, error) {
	tag := "without-redirect"
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New(tag)
		},
	}
	res, err := c.Do(req)
	if err != nil && !strings.Contains(err.Error(), tag) {
		return nil, err
	}
	return res, nil
}

func newTestServerAndNode(t *testing.T, ns mockNamesys) (*httptest.Server, *core.IpfsNode) {
	n, err := newNodeWithMockNamesys(ns)
	if err != nil {
		t.Fatal(err)
	}

	// need this variable here since we need to construct handler with
	// listener, and server with handler. yay cycles.
	dh := &delegatedHandler{}
	ts := httptest.NewServer(dh)

	dh.Handler, err = makeHandler(n,
		ts.Listener,
		IPNSHostnameOption(),
		GatewayOption(false),
	)
	if err != nil {
		t.Fatal(err)
	}

	return ts, n
}

func TestGatewayGet(t *testing.T) {
	ns := mockNamesys{}
	ts, n := newTestServerAndNode(t, ns)
	defer ts.Close()

	k, err := coreunix.Add(n, strings.NewReader("fnord"))
	if err != nil {
		t.Fatal(err)
	}
	ns["/ipns/example.com"] = path.FromString("/ipfs/" + k)

	t.Log(ts.URL)
	for _, test := range []struct {
		host   string
		path   string
		status int
		text   string
	}{
		{"localhost:5001", "/", http.StatusNotFound, "404 page not found\n"},
		{"localhost:5001", "/" + k, http.StatusNotFound, "404 page not found\n"},
		{"localhost:5001", "/ipfs/" + k, http.StatusOK, "fnord"},
		{"localhost:5001", "/ipns/nxdomain.example.com", http.StatusBadRequest, "Path Resolve error: " + namesys.ErrResolveFailed.Error()},
		{"localhost:5001", "/ipns/example.com", http.StatusOK, "fnord"},
		{"example.com", "/", http.StatusOK, "fnord"},
	} {
		var c http.Client
		r, err := http.NewRequest("GET", ts.URL+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Host = test.host
		resp, err := c.Do(r)

		urlstr := "http://" + test.host + test.path
		if err != nil {
			t.Errorf("error requesting %s: %s", urlstr, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != test.status {
			t.Errorf("got %d, expected %d from %s", resp.StatusCode, test.status, urlstr)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error reading response from %s: %s", urlstr, err)
		}
		if string(body) != test.text {
			t.Errorf("unexpected response body from %s: expected %q; got %q", urlstr, test.text, body)
			continue
		}
	}
}

func TestIPNSHostnameRedirect(t *testing.T) {
	ns := mockNamesys{}
	ts, n := newTestServerAndNode(t, ns)
	t.Logf("test server url: %s", ts.URL)
	defer ts.Close()

	// create /ipns/example.net/foo/index.html
	_, dagn1, err := coreunix.AddWrapped(n, strings.NewReader("_"), "_")
	if err != nil {
		t.Fatal(err)
	}
	_, dagn2, err := coreunix.AddWrapped(n, strings.NewReader("_"), "index.html")
	if err != nil {
		t.Fatal(err)
	}
	dagn1.AddNodeLink("foo", dagn2)
	if err != nil {
		t.Fatal(err)
	}

	err = n.DAG.AddRecursive(dagn1)
	if err != nil {
		t.Fatal(err)
	}

	k, err := dagn1.Key()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("k: %s\n", k)
	ns["/ipns/example.net"] = path.FromString("/ipfs/" + k.String())

	// make request to directory containing index.html
	req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.net"

	res, err := doWithoutRedirect(req)
	if err != nil {
		t.Fatal(err)
	}

	// expect 302 redirect to same path, but with trailing slash
	if res.StatusCode != 302 {
		t.Errorf("status is %d, expected 302", res.StatusCode)
	}
	hdr := res.Header["Location"]
	if len(hdr) < 1 {
		t.Errorf("location header not present")
	} else if hdr[0] != "/foo/" {
		t.Errorf("location header is %v, expected /foo/", hdr[0])
	}
}

func TestIPNSHostnameBacklinks(t *testing.T) {
	ns := mockNamesys{}
	ts, n := newTestServerAndNode(t, ns)
	t.Logf("test server url: %s", ts.URL)
	defer ts.Close()

	// create /ipns/example.net/foo/
	_, dagn1, err := coreunix.AddWrapped(n, strings.NewReader("1"), "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, dagn2, err := coreunix.AddWrapped(n, strings.NewReader("2"), "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, dagn3, err := coreunix.AddWrapped(n, strings.NewReader("3"), "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	dagn2.AddNodeLink("bar", dagn3)
	dagn1.AddNodeLink("foo", dagn2)
	if err != nil {
		t.Fatal(err)
	}

	err = n.DAG.AddRecursive(dagn1)
	if err != nil {
		t.Fatal(err)
	}

	k, err := dagn1.Key()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("k: %s\n", k)
	ns["/ipns/example.net"] = path.FromString("/ipfs/" + k.String())

	// make request to directory listing
	req, err := http.NewRequest("GET", ts.URL+"/foo/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.net"

	res, err := doWithoutRedirect(req)
	if err != nil {
		t.Fatal(err)
	}

	// expect correct backlinks
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response: %s", err)
	}
	s := string(body)
	t.Logf("body: %s\n", string(body))

	if !strings.Contains(s, "Index of /foo/") {
		t.Fatalf("expected a path in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/\">") {
		t.Fatalf("expected backlink in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/foo/file.txt\">") {
		t.Fatalf("expected file in directory listing")
	}

	// make request to directory listing
	req, err = http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.net"

	res, err = doWithoutRedirect(req)
	if err != nil {
		t.Fatal(err)
	}

	// expect correct backlinks
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response: %s", err)
	}
	s = string(body)
	t.Logf("body: %s\n", string(body))

	if !strings.Contains(s, "Index of /") {
		t.Fatalf("expected a path in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/\">") {
		t.Fatalf("expected backlink in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/file.txt\">") {
		t.Fatalf("expected file in directory listing")
	}

	// make request to directory listing
	req, err = http.NewRequest("GET", ts.URL+"/foo/bar/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.net"

	res, err = doWithoutRedirect(req)
	if err != nil {
		t.Fatal(err)
	}

	// expect correct backlinks
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response: %s", err)
	}
	s = string(body)
	t.Logf("body: %s\n", string(body))

	if !strings.Contains(s, "Index of /foo/bar/") {
		t.Fatalf("expected a path in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/foo/\">") {
		t.Fatalf("expected backlink in directory listing")
	}
	if !strings.Contains(s, "<a href=\"/foo/bar/file.txt\">") {
		t.Fatalf("expected file in directory listing")
	}
}