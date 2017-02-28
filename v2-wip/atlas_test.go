package atlas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type atlasServer struct {
	URL *url.URL

	t      *testing.T
	ln     net.Listener
	server *http.Server
}

type clientTestResp struct {
	RawPath string
	Host    string
	Header  http.Header
	Body    string
}

func newTestAtlasServer(t *testing.T) *atlasServer {
	hs := &atlasServer{t: t}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	hs.ln = ln

	hs.URL = &url.URL{
		Scheme: "http",
		Host:   ln.Addr().String(),
	}

	mux := http.NewServeMux()
	hs.setupRoutes(mux)

	// TODO: this should be using httptest.Server
	server := &http.Server{}
	server.Handler = mux
	hs.server = server
	go server.Serve(ln)

	return hs
}

func (hs *atlasServer) Stop() {
	hs.ln.Close()
}

func (hs *atlasServer) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/_json", hs.jsonHandler)
	mux.HandleFunc("/_error", hs.errorHandler)
	mux.HandleFunc("/_status/", hs.statusHandler)

	mux.HandleFunc("/_binstore/", hs.binstoreHandler)

	// add an endpoint for testing arbitrary requests
	mux.HandleFunc("/_test", hs.testHandler)
}

// testHandler echos the data sent from the client in a json object
func (hs *atlasServer) testHandler(w http.ResponseWriter, r *http.Request) {

	req := &clientTestResp{
		RawPath: r.URL.RawPath,
		Host:    r.Host,
		Header:  r.Header,
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// log this, since an error should fail the test anyway
		hs.t.Log("error reading body:", err)
	}

	req.Body = string(body)

	js, _ := json.Marshal(req)
	if err != nil {
		hs.t.Log("error marshaling req:", err)
	}

	w.Write(js)
}

func (hs *atlasServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	slice := strings.Split(r.URL.Path, "/")
	codeStr := slice[len(slice)-1]

	code, err := strconv.ParseInt(codeStr, 10, 32)
	if err != nil {
		hs.t.Fatal(err)
	}

	w.WriteHeader(int(code))
}

func (hs *atlasServer) errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(422)
	w.Header().Set("Content-Type", "application/vnd.api+json")
	fmt.Fprintf(w, `{"errors":[{"title":"bad","detail":"detail","code":"114","status":"400"}]}`)
}

func (hs *atlasServer) jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok": true}`)
}

func (hs *atlasServer) binstoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
}
