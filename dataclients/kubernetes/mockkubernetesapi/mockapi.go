/*
Package mockkubernetesapi provides a mock for the Kubernetes API supporting ingress related features.

TODO:
- delete option for different objects
- accept full spec documents
- accept multiple ingress spec documents
*/
package mockkubernetesapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/zalando/skipper/pathmux"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type MockAPI struct {
	mux       *pathmux.Tree
	sync      sync.RWMutex
	ingresses []interface{}
	services  map[string]map[string]interface{}
	endpoints map[string]map[string]interface{}
}

func New() *MockAPI {
	a := &MockAPI{
		services:  make(map[string]map[string]interface{}),
		endpoints: make(map[string]map[string]interface{}),
	}

	a.updateMux()
	return a
}

func getName(obj interface{}) (ns string, n string) {
	jsonObj, ok := obj.(map[string]interface{})
	if !ok {
		return
	}

	metaObj, ok := jsonObj["metadata"].(map[string]interface{})
	if !ok {
		return
	}

	ns, ok = metaObj["namespace"].(string)
	if !ok {
		return
	}

	n, ok = metaObj["name"].(string)
	if !ok {
		return
	}

	return
}

func getKind(obj interface{}) string {
	jsonObj, ok := obj.(map[string]interface{})
	if !ok {
		return ""
	}

	kind, _ := jsonObj["kind"].(string)
	return kind
}

func findIndexByName(list []interface{}, ns, n string) int {
	for i, item := range list {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		meta, ok := obj["metadata"].(map[string]interface{})
		if !ok {
			continue
		}

		if meta["namespace"] == ns && meta["name"] == n {
			return i
		}
	}

	return -1
}

func findByKind(list []interface{}, kind string) []interface{} {
	var ofKind []interface{}
	for _, obj := range list {
		if strings.ToLower(getKind(obj)) == strings.ToLower(kind) {
			ofKind = append(ofKind, obj)
		}
	}

	return ofKind
}

func deleteByName(specs []interface{}, ns, n string) []interface{} {
	i := findIndexByName(specs, ns, n)
	if i < 0 {
		return nil
	}

	specs[i], specs[len(specs)-1] = specs[len(specs)-1], nil
	return specs[:len(specs)-1]
}

func unmarshalYAMLs(b []byte) ([]interface{}, error) {
	dec := yaml.NewYAMLToJSONDecoder(bytes.NewBuffer(b))

	var result []interface{}
	for {
		var obj interface{}
		if err := dec.Decode(&obj); err == io.EOF {
			return result, nil
		} else if err != nil {
			return nil, err
		}

		result = append(result, obj)
	}
}

func (a *MockAPI) updateMux() error {
	if a.ingresses == nil {
		// never 404, see ServeHTTP
		a.ingresses = make([]interface{}, 0)
	}

	mux := &pathmux.Tree{}

	var err error
	add := func(path string, r interface{}) {
		if err != nil {
			return
		}

		err = mux.Add(path, r)
	}

	add("/api/v1/namespaces/:namespace/services/:name", a.services)
	add("/api/v1/namespaces/:namespace/endpoints/:name", a.endpoints)
	add("/apis/extensions/v1beta1/ingresses", a.ingresses)

	a.mux = mux
	return err
}

func mergeMap(l ...[]interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{})
	for i := range l {
		for j := range l[i] {
			ns, n := getName(l[i][j])
			if _, ok := m[ns]; !ok {
				m[ns] = make(map[string]interface{})
			}

			m[ns][n] = l[i][j]
		}
	}

	return m
}

func mapToList(m map[string]map[string]interface{}) []interface{} {
	var l []interface{}
	for _, ns := range m {
		for _, obj := range ns {
			l = append(l, obj)
		}
	}

	return l
}

func mergeList(l ...[]interface{}) []interface{} {
	unique := mergeMap(l...)
	return mapToList(unique)
}

func (a *MockAPI) loadMapped(field *map[string]map[string]interface{}, y string) error {
	obj, err := unmarshalYAMLs([]byte(y))
	if err != nil {
		return err
	}

	a.sync.Lock()
	defer a.sync.Unlock()
	*field = mergeMap(mapToList(*field), obj)
	return a.updateMux()
}

func (a *MockAPI) deleteFromMap(field map[string]map[string]interface{}, ns, n string) {
	a.sync.Lock()
	defer a.sync.Unlock()

	nsm, ok := field[ns]
	if !ok {
		return
	}

	delete(nsm, n)
	a.updateMux()
}

func (a *MockAPI) LoadServices(y string) error {
	return a.loadMapped(&a.services, y)
}

func (a *MockAPI) DeleteService(ns, n string) {
	a.deleteFromMap(a.services, ns, n)
}

func (a *MockAPI) LoadEndpoints(y string) error {
	return a.loadMapped(&a.endpoints, y)
}

func (a *MockAPI) DeleteEndpoint(ns, n string) {
	a.deleteFromMap(a.endpoints, ns, n)
}

func (a *MockAPI) LoadIngresses(y string) error {
	obj, err := unmarshalYAMLs([]byte(y))
	if err != nil {
		return err
	}

	a.sync.Lock()
	defer a.sync.Unlock()
	a.ingresses = mergeList(a.ingresses, obj)
	return a.updateMux()
}

func (a *MockAPI) DeleteIngress(ns, n string) {
	a.sync.Lock()
	defer a.sync.Unlock()
	a.ingresses = deleteByName(a.ingresses, ns, n)
	a.updateMux()
}

func (a *MockAPI) Load(doc string) error {
	specs, err := unmarshalYAMLs([]byte(doc))
	if err != nil {
		return err
	}

	services := findByKind(specs, "service")
	endpoints := findByKind(specs, "endpoint")
	ingresses := findByKind(specs, "ingress")

	a.sync.Lock()
	defer a.sync.Unlock()

	a.services = mergeMap(mapToList(a.services), services)
	a.endpoints = mergeMap(mapToList(a.endpoints), endpoints)
	a.ingresses = mergeList(a.ingresses, ingresses)

	a.updateMux()
	return nil
}

func (a *MockAPI) Reset(doc string) error {
	specs, err := unmarshalYAMLs([]byte(doc))
	if err != nil {
		return err
	}

	services := findByKind(specs, "service")
	endpoints := findByKind(specs, "endpoint")
	ingresses := findByKind(specs, "ingress")

	a.sync.Lock()
	defer a.sync.Unlock()

	a.services = mergeMap(services)
	a.endpoints = mergeMap(endpoints)
	a.ingresses = mergeList(ingresses)

	a.updateMux()
	return nil
}

func (a *MockAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.sync.RLock()
	rt, p := a.mux.Lookup(r.URL.Path)
	a.sync.RUnlock()

	if rt == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var d interface{}
	switch rtt := rt.(type) {
	case []interface{}:
		d = map[string]interface{}{
			"items": rt,
		}
	case map[string]map[string]interface{}:
		var ok bool
		d, ok = rtt[p["namespace"]][p["name"]]
		if !ok {
			log.Println("not found:", p["namespace"], p["name"])
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		log.Println("invalid data type")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.Println("marshal failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}
