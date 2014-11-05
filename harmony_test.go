package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
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

	mux.HandleFunc("/_binstore/", hs.binstoreHandler)

	mux.HandleFunc("/api/v1/authenticate", hs.authenticationHandler)

	mux.HandleFunc("/api/v1/artifacts/hashicorp/existing", hs.vagrantArtifactExistingHandler)
	mux.HandleFunc(
		"/api/v1/artifacts/hashicorp/existing/amazon-ami",
		hs.vagrantArtifactUploadHandler)
	mux.HandleFunc(
		"/api/v1/artifacts/hashicorp/existing1/amazon-ami/search",
		hs.vagrantArtifactSearchHandler1)
	mux.HandleFunc(
		"/api/v1/artifacts/hashicorp/existing2/amazon-ami/search",
		hs.vagrantArtifactSearchHandler2)

	mux.HandleFunc("/api/v1/vagrant/applications", hs.vagrantCreateAppHandler)
	mux.HandleFunc("/api/v1/vagrant/applications/", hs.vagrantCreateAppsHandler)
	mux.HandleFunc("/api/v1/vagrant/applications/hashicorp/existing/version", hs.vagrantUploadAppHandler)

	mux.HandleFunc("/api/v1/packer/build-configurations", hs.vagrantBCCreateHandler)
	mux.HandleFunc("/api/v1/packer/build-configurations/hashicorp/existing", hs.vagrantBCExistingHandler)
	mux.HandleFunc(
		"/api/v1/packer/build-configurations/hashicorp/existing/version",
		hs.vagrantBCCreateVersionHandler)
}

func (hs *harmonyServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	slice := strings.Split(r.URL.Path, "/")
	codeStr := slice[len(slice)-1]

	code, err := strconv.ParseInt(codeStr, 10, 32)
	if err != nil {
		hs.t.Fatal(err)
	}

	w.WriteHeader(int(code))
}

func (hs *harmonyServer) jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok": true}`)
}

func (hs *harmonyServer) authenticationHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		hs.t.Fatal(err)
	}

	login, password := r.Form["user[login]"][0], r.Form["user[password]"][0]

	if login == "sethloves" && password == "bacon" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
      {
        "token": "pX4AQ5vO7T-xJrxsnvlB0cfeF-tGUX-A-280LPxoryhDAbwmox7PKinMgA1F6R3BKaT"
      }
    `)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `
      {
        "errors": {
          "error": [
            "Bad login details"
          ]
        }
      }
    `)
	}
}

func (hs *harmonyServer) vagrantArtifactExistingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, `
	{
		"artifact": {
			"username": "hashicorp",
			"name": "existing",
			"tag": "hashicorp/existing"
		}
	}
	`)
}

func (hs *harmonyServer) vagrantArtifactSearchHandler1(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, `
	{
		"versions": [{
			"username": "hashicorp",
			"name": "existing",
			"tag": "hashicorp/existing"
		}]
	}
	`)
}

func (hs *harmonyServer) vagrantArtifactSearchHandler2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Form.Get("metadata.1.key") == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Form.Get("metadata.2.key") == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, `
	{
		"versions": [{
			"username": "hashicorp",
			"name": "existing",
			"tag": "hashicorp/existing"
		}]
	}
	`)
}

func (hs *harmonyServer) vagrantArtifactUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r.Body); err != nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if buf.Len() == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}

	uploadPath := hs.URL.String() + "/_binstore/"

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
	{
		"upload_path": "%s"
	}
	`, uploadPath)
}

func (hs *harmonyServer) vagrantBCCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var wrapper bcWrapper
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&wrapper); err != nil && err != io.EOF {
		hs.t.Fatal(err)
	}
	bc := wrapper.BuildConfig

	if bc.User != "hashicorp" {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (hs *harmonyServer) vagrantBCCreateVersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var wrapper bcCreateWrapper
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&wrapper); err != nil && err != io.EOF {
		hs.t.Fatal(err)
	}
	builds := wrapper.Version.Builds

	if len(builds) == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}

	uploadPath := hs.URL.String() + "/_binstore/"

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
	{
		"upload_path": "%s"
	}
	`, uploadPath)
}

func (hs *harmonyServer) vagrantBCExistingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, `
	{
		"build_configuration": {
			"username": "hashicorp",
			"name": "existing"
		}
	}
	`)
}

func (hs *harmonyServer) vagrantCreateAppHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var aw appWrapper
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&aw); err != nil && err != io.EOF {
		hs.t.Fatal(err)
	}
	app := aw.Application

	if app.User == "hashicorp" && app.Name == "existing" {
		w.WriteHeader(http.StatusConflict)
	} else {
		body, err := json.Marshal(&aw)
		if err != nil {
			hs.t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(body))
	}
}

func (hs *harmonyServer) vagrantCreateAppsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	split := strings.Split(r.RequestURI, "/")
	parts := split[len(split)-2:]
	user, name := parts[0], parts[1]

	if user == "hashicorp" && name == "existing" {
		body, err := json.Marshal(&appWrapper{&App{
			User: "hashicorp",
			Name: "existing",
		}})
		if err != nil {
			hs.t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(body))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (hs *harmonyServer) vagrantUploadAppHandler(w http.ResponseWriter, r *http.Request) {
	u := *hs.URL
	u.Path = path.Join(u.Path, "_binstore/630e42d9-2364-2412-4121-18266770468e")

	body, err := json.Marshal(&appVersion{
		UploadPath: &u,
		Token:      "630e42d9-2364-2412-4121-18266770468e",
		Version:    125,
	})
	if err != nil {
		hs.t.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(body))
}

func (hs *harmonyServer) binstoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
}
