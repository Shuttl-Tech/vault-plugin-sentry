package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

const AuthToken = "local-auth-token"

func main() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	handler := mux.NewRouter()
	registerRoutes(handler)

	srv := &http.Server{
		Handler: handler,
		Addr:    "127.0.0.1:" + os.Args[1],
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("http server shut down with error %s", err)
		}
	}()

	<-exit
	srv.Close()
}

func registerRoutes(h *mux.Router) {
	p := apiWithDefaultData()

	h.HandleFunc("/organizations/{org}", Auth(p.HandleGet)).Methods(http.MethodGet)
	h.HandleFunc("/projects/{org}/{project}", Auth(p.HandleGet)).Methods(http.MethodGet)
	h.HandleFunc("/teams/{org}/{team}/projects", Auth(p.HandleCreateProject)).Methods(http.MethodPost)
	h.HandleFunc("/projects/{org}/{project}/keys", Auth(p.HandleGet)).Methods(http.MethodGet)
	h.HandleFunc("/projects/{org}/{project}/keys", Auth(p.HandleCreateKey)).Methods(http.MethodPost)

	h.HandleFunc("/_dump", p.Dump)
}

type api struct {
	data map[string]string
}

func apiWithDefaultData() *api {
	return &api{
		data: map[string]string{
			"organizations/valid-test-org":               fmt.Sprintf(getOrgResp, "valid-test-org", "Valid Test Org"),
			"projects/valid-test-org/valid-test-project": fmt.Sprintf(getProjectResp, "valid-test-project", "Valid Test Project"),
			"projects/valid-test-org/valid-test-project/keys": fmt.Sprintf(getKeysResp, "test-key-1", "Test Key",
				"test-key-1", "test-keys-1", "test-key-1", "test-key-1",
				"test-key-1", "test-key-1", "test-key-1", "test-key-1", "test-key-1"),
		},
	}
}

func Auth(fn func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Printf("new request received for %s", req.URL)
		auth := req.Header.Get("Authorization")
		token := strings.ReplaceAll(auth, "Bearer ", "")
		if token != AuthToken {
			resp.WriteHeader(http.StatusUnauthorized)
			resp.Write([]byte(fmt.Sprintf("invalid api token: %q", token)))
			return
		}

		log.Printf("handling request for %s", req.URL)
		fn(resp, req)
	}
}

func (h *api) mkey(req *http.Request) string {
	return strings.Trim(req.URL.Path, "/")
}

func (h *api) Dump(resp http.ResponseWriter, req *http.Request) {
	err := json.NewEncoder(resp).Encode(h.data)
	if err != nil {
		log.Printf("failed to dump state. %s", err)
	}
}

func (h *api) HandleGet(resp http.ResponseWriter, req *http.Request) {
	if v, ok := h.data[h.mkey(req)]; ok {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(v))
	} else {
		resp.WriteHeader(http.StatusNotFound)
	}
}

func (h *api) HandleCreateProject(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	data := &struct {
		Name string `json:"name"`
	}{}

	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	project := fmt.Sprintf(getProjectResp, data.Name, data.Name)
	h.data[fmt.Sprintf("projects/%s/%s", vars["org"], data.Name)] = project

	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(project))
}

func (h *api) HandleCreateKey(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	data := &struct {
		Name string `json:"name"`
	}{}

	err := json.NewDecoder(req.Body).Decode(data)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	key := fmt.Sprintf(singleKey, data.Name, data.Name, data.Name)
	h.data[fmt.Sprintf("projects/%s/%s/keys", vars["org"], vars["project"])] = fmt.Sprintf("[%s]", key)

	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(key))
}
