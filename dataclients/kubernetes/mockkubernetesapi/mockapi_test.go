package mockkubernetesapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

const serviceYAML = `
metadata:
  namespace: test-namespace
  name: test-service
spec:
  clusterIP: 10.0.0.1
  ports:
  - name: http
    port: 80
    targetPort: 80
`

const endpointsYAML = `
subsets:
  addresses:
    - ip: 10.0.1.1
    - ip: 10.0.1.2
  ports:
    - port: 80
    - port: 80
`

const ingressYAML = `
- metadata:
    namespace: test-namespace
    name: test-ingress-1
  spec:
    rules:
      - host: www.example.org
        http:
          paths:
            - path: /
              backend:
                serviceName: test-service
                servicePort: 80
`

func Test(t *testing.T) {
	mapi := New()
	s := httptest.NewServer(mapi)
	defer s.Close()

	const (
		ns          = "test-namespace"
		serviceName = "test-service"
	)

	if err := mapi.SetServiceYAML(ns, serviceName, serviceYAML); err != nil {
		t.Fatal(err)
	}

	if err := mapi.SetEndpointsYAML(ns, serviceName, endpointsYAML); err != nil {
		t.Fatal(err)
	}

	if err := mapi.SetIngressYAML(ingressYAML); err != nil {
		t.Fatal(err)
	}

	var ingressBuf, serviceBuf, endpointsBuf bytes.Buffer
	for _, current := range []struct {
		buf *bytes.Buffer
		uri string
	}{
		{&ingressBuf, "/apis/extensions/v1beta1/ingresses"},
		{&serviceBuf, fmt.Sprintf("/api/v1/namespaces/%s/services/%s", ns, serviceName)},
		{&endpointsBuf, fmt.Sprintf("/api/v1/namespaces/%s/endpoints/%s", ns, serviceName)},
	} {
		func() {
			rsp, err := http.Get(s.URL + current.uri)
			if err != nil {
				t.Fatal(err)
			}

			defer rsp.Body.Close()

			if rsp.StatusCode != http.StatusOK {
				t.Fatal("invalid status code:", rsp.StatusCode)
			}

			if _, err = io.Copy(current.buf, rsp.Body); err != nil {
				t.Fatal(err)
			}
		}()
	}

	for _, current := range []struct {
		buf  bytes.Buffer
		yaml string
	}{
		{ingressBuf, ingressYAML},
		{serviceBuf, serviceYAML},
		{endpointsBuf, endpointsYAML},
	} {
		var obj interface{}
		if err := json.Unmarshal(current.buf.Bytes(), &obj); err != nil {
			t.Fatal(err)
		}

		var objCheck interface{}
		if err := yaml.Unmarshal([]byte(current.yaml), &objCheck); err != nil {
			t.Fatal(err)
		}

		objCheck, err := Jsonify(objCheck)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(obj, objCheck) {
			t.Fatal("failed to received the right API data")
		}
	}
}
