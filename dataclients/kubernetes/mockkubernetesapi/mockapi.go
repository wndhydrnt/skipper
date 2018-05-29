package mockkubernetesapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/zalando/skipper/pathmux"
	"gopkg.in/yaml.v2"
)

type mockapi struct {
	mux       *pathmux.Tree
	sync      sync.RWMutex
	ingress   []byte
	services  map[string]map[string][]byte
	endpoints map[string]map[string][]byte
}

func New() *mockapi {
	a := &mockapi{
		services:  make(map[string]map[string][]byte),
		endpoints: make(map[string]map[string][]byte),
	}

	a.updateMux()
	return a
}

func Jsonify(obj interface{}) (interface{}, error) {
	switch objt := obj.(type) {
	case map[interface{}]interface{}:
		ms := make(map[string]interface{})
		for k, v := range objt {
			switch kt := k.(type) {
			case string:
				v, err := Jsonify(v)
				if err != nil {
					return nil, err
				}

				ms[kt] = v
			default:
				return nil, errors.New("unaccepted key type")
			}
		}

		return ms, nil
	case []interface{}:
		for i := range objt {
			v, err := Jsonify(objt[i])
			if err != nil {
				return nil, err
			}

			objt[i] = v
		}

		return objt, nil
	case int:
		return float64(objt), nil
	default:
		return obj, nil
	}
}

func yamlToJSON(y []byte) ([]byte, error) {
	var obj interface{}
	if err := yaml.Unmarshal([]byte(y), &obj); err != nil {
		return nil, err
	}

	obj, err := Jsonify(obj)
	if err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

func (a *mockapi) updateMux() error {
	mux := &pathmux.Tree{}

	var err error
	add := func(path string, res interface{}) {
		if err != nil {
			return
		}

		err = mux.Add(path, res)
	}

	add("/apis/extensions/v1beta1/ingresses", a.ingress)
	add("/api/v1/namespaces/:namespace/services/:name", a.services)
	add("/api/v1/namespaces/:namespace/endpoints/:name", a.endpoints)

	a.mux = mux
	return err
}

func (a *mockapi) setServiceJSON(m map[string]map[string][]byte, ns, name string, j []byte) error {
	a.sync.Lock()
	defer a.sync.Unlock()
	if _, ok := m[ns]; !ok {
		m[ns] = make(map[string][]byte)
	}

	m[ns][name] = []byte(j)
	return a.updateMux()
}

func (a *mockapi) setServiceYAML(m map[string]map[string][]byte, ns, name string, y []byte) error {
	j, err := yamlToJSON(y)
	if err != nil {
		return err
	}

	return a.setServiceJSON(m, ns, name, j)
}

func (a *mockapi) SetServiceJSON(namespace, name, j string) error {
	return a.setServiceJSON(a.services, namespace, name, []byte(j))
}

func (a *mockapi) SetServiceYAML(namespace, name, y string) error {
	return a.setServiceYAML(a.services, namespace, name, []byte(y))
}

func (a *mockapi) SetEndpointsJSON(namespace, name, j string) error {
	return a.setServiceJSON(a.endpoints, namespace, name, []byte(j))
}

func (a *mockapi) SetEndpointsYAML(namespace, name, y string) error {
	return a.setServiceYAML(a.endpoints, namespace, name, []byte(y))
}

func (a *mockapi) setIngressJSON(j []byte) error {
	a.sync.Lock()
	defer a.sync.Unlock()
	a.ingress = j
	return a.updateMux()
}

func (a *mockapi) SetIngressJSON(j string) error {
	return a.setIngressJSON([]byte(j))
}

func (a *mockapi) SetIngressYAML(y string) error {
	j, err := yamlToJSON([]byte(y))
	if err != nil {
		return err
	}

	return a.setIngressJSON(j)
}

func (a *mockapi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	a.sync.RLock()
	res, p := a.mux.Lookup(r.URL.Path)
	a.sync.RUnlock()

	if res == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch rest := res.(type) {
	case []byte:
		w.Write(rest)
	case map[string]map[string][]byte:
		d, ok := rest[p["namespace"]][p["name"]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(d)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
