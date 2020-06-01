package backend

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/logical"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type testSentryHandler struct {
	mux *http.ServeMux
	url string
}

var localSentry = &testSentryHandler{
	mux: http.NewServeMux(),
}

func TestMain(m *testing.M) {
	log.Println("====> Starting test server")

	testSrv := httptest.NewServer(localSentry)
	localSentry.url = testSrv.URL + "/"
	result := m.Run()
	testSrv.Close()

	os.Exit(result)
}

func (m *testSentryHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	h, p := m.mux.Handler(req)

	if p == "" {
		panic(fmt.Sprintf("====> handler is not registered for %q", req.URL))
	}

	log.Printf("====> handling request for %q", p)
	h.ServeHTTP(resp, req)
}

func (m *testSentryHandler) handleStatic(route string, code int, content string) {
	log.Printf("====> registering static handler for %s -> %d", route, code)
	m.mux.HandleFunc(route, func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(code)
		resp.Write([]byte(content))
	})
}

func testGetBackend(t *testing.T) logical.Backend {
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}
	b, err := Factory(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to initialize backend factory. %s", err)
	}

	return b
}
