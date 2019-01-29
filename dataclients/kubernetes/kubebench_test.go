package kubernetes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/zalando/skipper/eskip"
)

type benchAPI struct {
	bench     *testing.B
	services  services
	ingresses *ingressList
	endpoints endpoints
	server    *httptest.Server
}

func (api *benchAPI) getEndpoints(uri string) endpoint {
	var ep endpoint
	if m := endpointURIRx.FindAllStringSubmatch(uri, -1); len(m) != 0 {
		ns, n := m[0][1], m[0][2]
		ep = api.endpoints[ns][n]
	}

	return ep
}

func (api *benchAPI) getTestService(uri string) (*service, bool) {
	if m := serviceURIRx.FindAllStringSubmatch(uri, -1); len(m) != 0 {
		ns, n := m[0][1], m[0][2]
		return api.services[ns][n], true
	}

	return nil, false
}

func (api *benchAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == ingressesClusterURI {
		if err := respondJSON(w, api.ingresses); err != nil {
			api.bench.Error(err)
		}

		return
	}

	if endpointURIRx.MatchString(r.URL.Path) {
		ep := api.getEndpoints(r.URL.Path)
		if err := respondJSON(w, ep); err != nil {
			api.bench.Error(err)
		}

		return
	}

	s, ok := api.getTestService(r.URL.Path)
	if !ok {
		s = &service{}
	}

	if err := respondJSON(w, s); err != nil {
		api.bench.Error(err)
	}
}

func (api *benchAPI) Close() {
	api.server.Close()
}

func newBenchAPI(b *testing.B, s services, i *ingressList) *benchAPI {
	return newBenchAPIWithEndpoints(b, s, i, nil)
}

func newBenchAPIWithEndpoints(b *testing.B, s services, i *ingressList, e endpoints) *benchAPI {
	api := &benchAPI{
		bench:     b,
		services:  s,
		ingresses: i,
		endpoints: e,
	}

	api.server = httptest.NewServer(api)
	return api
}

func benchIngresses(n int) []*ingressItem {
	result := make([]*ingressItem, 0, n)

	for i := 0; i < n; i++ {
		ing := testIngressSimple(
			"namespace1",
			"name"+strconv.Itoa(i),
			"service1",
			backendPort{"port1"},
		)
		result = append(result, ing)
	}

	return result
}

func BenchmarkLoadUpdateEastWest(b *testing.B) {
	api := newBenchAPI(b, nil, &ingressList{})
	api.services = testServices()
	api.ingresses.Items = benchIngresses(1)
	dc, err := New(Options{
		KubernetesURL:            api.server.URL,
		KubernetesEnableEastWest: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	_, err = dc.LoadAll()
	if err != nil {
		b.Fatal(err)
	}

	for n := 1; n < b.N; n++ {
		api.ingresses.Items = benchIngresses(n * 50)
		_, _, err := dc.LoadUpdate()
		if err != nil {
			b.Error("failed to fail")
		}
		//b.Logf("Loaded %d routes, deleted %d routes", len(r), len(d))
	}

	dc.Close()
	api.Close()
}

func BenchmarkCreateEastWestRoutes(b *testing.B) {
	routes := make([]*eskip.Route, 0)
	ns := "default"
	nameFmt := "sugarcane-test-pgieschen-foo-bar-commodity-groups-id-admin-skipper%d"
	for i := 0; i < 2000; i++ {
		r := &eskip.Route{
			Id:      routeID(ns, fmt.Sprintf(nameFmt, i), "", "", "", 0),
			Backend: "http://10.2.12.52:8080/",
		}
		routes = append(routes, r)
	}

	for n := 0; n < b.N; n++ {
		createEastWestRoutes(".ingress.cluster.local", "anothername"+strconv.Itoa(n), "default", routes)
	}
}
