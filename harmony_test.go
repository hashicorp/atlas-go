package harmony

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type harmonyServer struct {
	URL *url.URL

	t      *testing.T
	ln     net.Listener
	server http.Server
}

func newTestHarmonyServer(t *testing.T) *harmonyServer {
	hs := &harmonyServer{t: t}

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

	var server http.Server
	server.Handler = mux
	hs.server = server
	go server.Serve(ln)

	return hs
}

func (hs *harmonyServer) Stop() {
	hs.ln.Close()
}

func (hs *harmonyServer) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/_json", hs.jsonHandler)
	mux.HandleFunc("/_status/", hs.statusHandler)
}

func (hs *harmonyServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	slice := strings.Split(r.URL.Path, "/")
	codeStr := slice[len(slice)-1]

	code, err := strconv.ParseInt(codeStr, 10, 32)
	if err != nil {
		hs.t.Fatal(err)
	}

	w.WriteHeader(int(code))
	fmt.Fprintf(w, "Status code: %d", code)
}

func (hs *harmonyServer) jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok": true}`)
}
