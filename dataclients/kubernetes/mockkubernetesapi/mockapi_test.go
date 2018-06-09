package mockkubernetesapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

const (
	servicesURIFmt  = "/api/v1/namespaces/%s/services/%s"
	endpointsURIFmt = "/api/v1/namespaces/%s/endpoints/%s"
	ingressesURI    = "/apis/extensions/v1beta1/ingresses"

	testNamespace   = "test-namespace"
	initialService1 = "test-service-1"
	initialService2 = "test-service-2"
	initialService3 = "test-service-3"
	initialIngress1 = "test-ingress-1"
)

const initialServices = `
kind: Service
metadata:
  namespace: test-namespace
  name: test-service-1
spec:
  clusterIP: 10.0.0.1
  ports:
  - name: http
    port: 80
    targetPort: 80
---
kind: Service
metadata:
  namespace: test-namespace
  name: test-service-2
spec:
  clusterIP: 10.0.0.2
  ports:
  - name: http
    port: 80
    targetPort: 80
`

const initialEndpoints = `
kind: Endpoint
metadata:
  namespace: test-namespace
  name: test-service-1
subsets:
  addresses:
    - ip: 10.0.1.1
    - ip: 10.0.1.2
  ports:
    - port: 80
    - port: 80
---
kind: Endpoint
metadata:
  namespace: test-namespace
  name: test-service-2
subsets:
  addresses:
    - ip: 10.0.1.3
    - ip: 10.0.1.4
  ports:
    - port: 80
    - port: 80
`

const initialIngresses = `
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-1
spec:
  rules:
    - host: www.example.org
      http:
        paths:
          - path: /
            backend:
              serviceName: test-service-1
              servicePort: 80
---
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-2
spec:
  rules:
    - host: api.example.org
      http:
        paths:
          - path: /api
            backend:
              serviceName: test-service-2
              servicePort: 80
`

const updatedServices = `
kind: Service
metadata:
  namespace: test-namespace
  name: test-service-1
spec:
  clusterIP: 10.0.0.1
  ports:
  - name: http
    port: 8080
    targetPort: 8080
---
kind: Service
metadata:
  namespace: test-namespace
  name: test-service-3
spec:
  clusterIP: 10.0.0.3
  ports:
  - name: http
    port: 80
    targetPort: 80
`

const updatedEndpoints = `
kind: Endpoint
metadata:
  namespace: test-namespace
  name: test-service-2
subsets:
  addresses:
    - ip: 10.0.1.1
    - ip: 10.0.1.2
  ports:
    - port: 8080
    - port: 8080
---
kind: Endpoint
metadata:
  namespace: test-namespace
  name: test-service-3
subsets:
  addresses:
    - ip: 10.0.1.5
    - ip: 10.0.1.6
  ports:
    - port: 80
    - port: 80
`

const updatedIngresses = `
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-1
spec:
  rules:
    - host: www.example.org
      http:
        paths:
          - path: /
            backend:
              serviceName: test-service-1
              servicePort: 80
---
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-3
spec:
  rules:
    - host: api.example.org
      http:
        paths:
          - path: /test
            backend:
              serviceName: test-service-3
              servicePort: 7272
`

const singleDoc = `
kind: Service
metadata:
  namespace: test-namespace
  name: test-service-1
spec:
  clusterIP: 10.0.0.1
  ports:
  - name: http
    port: 80
    targetPort: 80
---
kind: Endpoint
metadata:
  namespace: test-namespace
  name: test-service-1
subsets:
  addresses:
    - ip: 10.0.1.1
    - ip: 10.0.1.2
  ports:
    - port: 80
    - port: 80
---
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-1
spec:
  rules:
    - host: www.example.org
      http:
        paths:
          - path: /
            backend:
              serviceName: test-service-1
              servicePort: 80
`

const singleIngress = `
kind: Ingress
metadata:
  namespace: test-namespace
  name: test-ingress-1
spec:
  rules:
    - host: www.example.org
      http:
        paths:
          - path: /
            backend:
              serviceName: test-service-1
              servicePort: 80
`

type sortMeta []interface{}

func (sm sortMeta) Len() int      { return len(sm) }
func (sm sortMeta) Swap(i, j int) { sm[i], sm[j] = sm[j], sm[i] }

func (sm sortMeta) Less(i, j int) bool {
	nsi, ni := getName(sm[i])
	nsj, nj := getName(sm[j])
	return fmt.Sprintf("%s/%s", nsi, ni) < fmt.Sprintf("%s/%s", nsj, nj)
}

func initAPI() (*MockAPI, error) {
	mapi := New()

	var err error
	load := func(f func(string) error, doc string) {
		if err != nil {
			return
		}

		err = f(doc)
	}

	load(mapi.LoadServices, initialServices)
	load(mapi.LoadEndpoints, initialEndpoints)
	load(mapi.LoadIngresses, initialIngresses)

	return mapi, err
}

func apiGet(urlBase, uriFmt string, params ...interface{}) ([]byte, error) {
	rsp, err := http.Get(urlBase + fmt.Sprintf(uriFmt, params...))
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rsp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func apiFindNot(urlBase, uriFmt string, params ...interface{}) error {
	rsp, err := http.Get(urlBase + fmt.Sprintf(uriFmt, params...))
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("invalid status code: %d", rsp.StatusCode)
	}

	return nil
}

func sortByMeta(list []interface{}) []interface{} {
	sm := make(sortMeta, len(list))
	copy(sm, list)
	sort.Sort(sm)

	result := make([]interface{}, len(list))
	copy(result, sm)
	return result
}

func findByName(list []interface{}, ns, n string) (interface{}, bool) {
	i := findIndexByName(list, ns, n)
	if i < 0 {
		return nil, false
	}

	return list[i], true
}

func getSpecs(sources []string) ([]interface{}, error) {
	var specs [][]interface{}
	for _, s := range sources {
		si, err := unmarshalYAMLs([]byte(s))
		if err != nil {
			return nil, err
		}

		specs = append(specs, si)
	}

	return mergeList(specs...), nil
}

func getExpectedSpec(sources []string, n string) (interface{}, bool, error) {
	specs, err := getSpecs(sources)
	if err != nil {
		return nil, false, err
	}

	spec, ok := findByName(specs, testNamespace, n)
	return spec, ok, nil
}

func getExpectedList(sources []string) ([]interface{}, error) {
	specs, err := getSpecs(sources)
	if err != nil {
		return nil, err
	}

	return sortByMeta(specs), nil
}

func removeSpec(name string, sources ...string) ([]interface{}, error) {
	specs, err := getSpecs(sources)
	if err != nil {
		return nil, err
	}

	specs = deleteByName(specs, testNamespace, name)
	specs = sortByMeta(specs)
	return specs, nil
}

func testSpec(t *testing.T, url, format, name string, specSources ...string) {
	b, err := apiGet(url, format, testNamespace, name)
	if err != nil {
		t.Fatal(err)
	}

	var obj interface{}
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatal(err)
	}

	expectedSpec, ok, err := getExpectedSpec(specSources, name)
	if err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("initial spec not found")
	}

	if !reflect.DeepEqual(obj, expectedSpec) {
		t.Error("invalid response received")
		t.Log(pretty.Compare(obj, expectedSpec))
	}
}

func testSpecMissing(t *testing.T, baseURL, format, name string) {
	if err := apiFindNot(baseURL, format, testNamespace, name); err != nil {
		t.Fatal(err)
	}
}

func testListSpecs(t *testing.T, baseURL, uri string, expectedSpecs []interface{}) {
	b, err := apiGet(baseURL, uri)
	if err != nil {
		t.Fatal(err)
	}

	var obj interface{}
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatal(err)
	}

	top, ok := obj.(map[string]interface{})
	if !ok {
		t.Error("invalid format received")
		t.Log(pretty.Compare(obj, expectedSpecs))
		return
	}

	list, ok := top["items"].([]interface{})
	if !ok {
		t.Error("invalid format received")
		t.Log(pretty.Compare(obj, expectedSpecs))
		return
	}

	top["items"] = sortByMeta(list)

	expectedWrappedSpecs := map[string]interface{}{"items": expectedSpecs}
	if !reflect.DeepEqual(obj, expectedWrappedSpecs) {
		t.Error("invalid response received")
		t.Log(pretty.Compare(obj, expectedWrappedSpecs))
	}
}

func testList(t *testing.T, baseURL, uri string, specSources ...string) {
	expectedSpecs, err := getExpectedList(specSources)
	if err != nil {
		t.Fatal("initial service not found")
	}

	testListSpecs(t, baseURL, uri, expectedSpecs)
}

func Test(t *testing.T) {
	t.Run("initial", func(t *testing.T) {
		api, err := initAPI()
		if err != nil {
			t.Fatal(err)
		}

		s := httptest.NewServer(api)
		defer s.Close()

		testSpec(t, s.URL, servicesURIFmt, initialService1, initialServices)
		testSpec(t, s.URL, servicesURIFmt, initialService2, initialServices)
		testSpec(t, s.URL, endpointsURIFmt, initialService1, initialEndpoints)
		testSpec(t, s.URL, endpointsURIFmt, initialService2, initialEndpoints)
		testList(t, s.URL, ingressesURI, initialIngresses)
	})

	t.Run("append and overwrite", func(t *testing.T) {
		api, err := initAPI()
		if err != nil {
			t.Fatal(err)
		}

		s := httptest.NewServer(api)
		defer s.Close()

		api.LoadServices(updatedServices)
		api.LoadEndpoints(updatedEndpoints)
		api.LoadIngresses(updatedIngresses)

		testSpec(t, s.URL, servicesURIFmt, initialService1, initialServices, updatedServices)
		testSpec(t, s.URL, servicesURIFmt, initialService2, initialServices, updatedServices)
		testSpec(t, s.URL, servicesURIFmt, initialService3, initialServices, updatedServices)
		testSpec(t, s.URL, endpointsURIFmt, initialService1, initialEndpoints, updatedEndpoints)
		testSpec(t, s.URL, endpointsURIFmt, initialService2, initialEndpoints, updatedEndpoints)
		testSpec(t, s.URL, endpointsURIFmt, initialService3, initialEndpoints, updatedEndpoints)
		testList(t, s.URL, ingressesURI, initialIngresses, updatedIngresses)
	})

	t.Run("delete", func(t *testing.T) {
		api, err := initAPI()
		if err != nil {
			t.Fatal(err)
		}

		s := httptest.NewServer(api)
		defer s.Close()

		api.DeleteService(testNamespace, initialService1)
		api.DeleteEndpoint(testNamespace, initialService1)
		api.DeleteIngress(testNamespace, initialIngress1)

		expectedIngresses, err := removeSpec(initialIngress1, initialIngresses)
		if err != nil {
			t.Fatal(err)
		}

		testSpecMissing(t, s.URL, servicesURIFmt, initialService1)
		testSpec(t, s.URL, servicesURIFmt, initialService2, initialServices)
		testSpecMissing(t, s.URL, endpointsURIFmt, initialService1)
		testSpec(t, s.URL, endpointsURIFmt, initialService2, initialEndpoints)
		testListSpecs(t, s.URL, ingressesURI, expectedIngresses)
	})
}

func TestSingleDoc(t *testing.T) {
	api := New()
	s := httptest.NewServer(api)
	defer s.Close()

	api.Load(singleDoc)

	testSpec(t, s.URL, servicesURIFmt, initialService1, initialServices)
	// testSpec(t, s.URL, endpointsURIFmt, initialService1, initialEndpoints)
	// testList(t, s.URL, ingressesURI, singleIngress)
}

func TestReset(t *testing.T) {
	api, err := initAPI()
	if err != nil {
		t.Fatal(err)
	}

	s := httptest.NewServer(api)
	defer s.Close()

	api.Reset(singleDoc)

	testSpec(t, s.URL, servicesURIFmt, initialService1, initialServices)
	testSpecMissing(t, s.URL, servicesURIFmt, initialService2)
	testSpec(t, s.URL, endpointsURIFmt, initialService1, initialEndpoints)
	testSpecMissing(t, s.URL, endpointsURIFmt, initialService2)
	testList(t, s.URL, ingressesURI, singleIngress)
}
