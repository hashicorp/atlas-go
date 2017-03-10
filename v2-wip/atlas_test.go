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

	mux.HandleFunc("/api/v2/states/150-include-versions", hs.stateSingleIncludeVersionsHandler)
	mux.HandleFunc("/api/v2/state-versions/424", hs.stateVersionsSingleHandler)

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

func (hs *atlasServer) stateSingleIncludeVersionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `
{
  "data": {
    "id": "150",
    "type": "states",
    "links": {
      "self": "https://atlas.example.com/api/v2/states/150"
    },
    "attributes": {
      "created-at": "2017-03-10T00:51:19.359Z",
      "updated-at": "2017-03-10T00:52:44.233Z"
    },
    "relationships": {
      "configuration": {
        "links": {
          "self": "https://atlas.example.com/api/v2/states/150/relationships/configuration",
          "related": "https://atlas.example.com/api/v2/states/150/configuration"
        }
      },
      "environment": {
        "links": {
          "self": "https://atlas.example.com/api/v2/states/150/relationships/environment",
          "related": "https://atlas.example.com/api/v2/states/150/environment"
        }
      },
      "tf-vars": {
        "links": {
          "self": "https://atlas.example.com/api/v2/states/150/relationships/tf-vars",
          "related": "https://atlas.example.com/api/v2/states/150/tf-vars"
        }
      },
      "versions": {
        "links": {
          "self": "https://atlas.example.com/api/v2/states/150/relationships/versions",
          "related": "https://atlas.example.com/api/v2/states/150/versions"
        },
        "data": [
          {
            "type": "state-versions",
            "id": "424"
          }
        ]
      },
      "head-version": {
        "links": {
          "self": "https://atlas.example.com/api/v2/states/150/relationships/head-version",
          "related": "https://atlas.example.com/api/v2/states/150/head-version"
        }
      }
    }
  },
  "included": [
    {
      "id": "424",
      "type": "state-versions",
      "links": {
        "self": "https://atlas.example.com/api/v2/state-versions/424"
      },
      "attributes": {
        "created-at": "2017-03-10T00:52:44.225Z",
        "updated-at": "2017-03-10T00:52:44.545Z",
        "serial": 3,
        "tfstate-file": "{\"version\": 3}",
        "version": 1
      },
      "relationships": {
        "state": {
          "links": {
            "self": "https://atlas.example.com/api/v2/state-versions/424/relationships/state",
            "related": "https://atlas.example.com/api/v2/state-versions/424/state"
          }
        },
        "run": {
          "links": {
            "self": "https://atlas.example.com/api/v2/state-versions/424/relationships/run",
            "related": "https://atlas.example.com/api/v2/state-versions/424/run"
          }
        }
      }
    }
  ]
}
`)
}

func (hs *atlasServer) stateVersionsSingleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `
{
  "data": {
    "id": "424",
    "type": "state-versions",
    "links": {
      "self": "https://atlas.example.com/api/v2/state-versions/424"
    },
    "attributes": {
      "created-at": "2017-03-10T00:52:44.225Z",
      "updated-at": "2017-03-10T00:52:44.545Z",
      "serial": 3,
      "tfstate-file": "{\"version\": 3}",
      "version": 1
    },
    "relationships": {
      "state": {
        "links": {
          "self": "https://atlas.example.com/api/v2/state-versions/424/relationships/state",
          "related": "https://atlas.example.com/api/v2/state-versions/424/state"
        }
      },
      "run": {
        "links": {
          "self": "https://atlas.example.com/api/v2/state-versions/424/relationships/run",
          "related": "https://atlas.example.com/api/v2/state-versions/424/run"
        }
      }
    }
  }
}
`)
}
