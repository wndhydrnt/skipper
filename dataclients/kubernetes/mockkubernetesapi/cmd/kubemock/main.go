package main

// TODO:
// - monitor spec dir for changes

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zalando/skipper/dataclients/kubernetes/mockkubernetesapi"
	"gopkg.in/yaml.v2"
)

func getKind(obj map[string]interface{}) string {
	kind, ok := obj["kind"].(string)
	if ok {
		return kind
	}

	if _, ok := obj["subsets"]; ok {
		return "endpoints"
	}

	return ""
}

func getStringMeta(obj map[string]interface{}, key string) string {
	m, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}

	s, _ := m[key].(string)
	return s
}

func main() {
	var specDir, address, tlsCertPath, tlsKeyPath string
	flag.StringVar(&specDir, "specs", ".", "directory containing the Kubernetes specs")
	flag.StringVar(&address, "address", ":8080", "address where the mock API will listen on")
	flag.StringVar(&tlsCertPath, "tls-cert", "", "path to TLS certificate")
	flag.StringVar(&tlsKeyPath, "tls-key", "", "path to TLS keyificate")
	flag.Parse()

	api := mockkubernetesapi.New()

	filepath.Walk(specDir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			log.Println(err)
			return nil
		}

		defer f.Close()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, f); err != nil {
			log.Println(err)
			return nil
		}

		var (
			obj    interface{}
			isYAML bool
		)

		if err := yaml.Unmarshal(buf.Bytes(), &obj); err != nil {
			if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
				log.Println("neither YAML or JSON", p)
				return nil
			}
		} else {
			isYAML = true
		}

		obj, err = mockkubernetesapi.Jsonify(obj)
		if err != nil {
			log.Println(err)
			return nil
		}

		var (
			kind, ns, name string
			ingressPath    string
			hasIngress     bool
		)

		switch objt := obj.(type) {
		case map[string]interface{}:
			kind = getKind(objt)
			ns = getStringMeta(objt, "namespace")
			name = getStringMeta(objt, "name")
		default:
			return nil
		}

		switch kind {
		case "ingress":
			if hasIngress {
				log.Printf(
					"only one ingress is allowed, found %s and %s\n",
					ingressPath,
					p,
				)

				return nil
			}

			hasIngress = true
			ingressPath = p
			if isYAML {
				api.SetIngressYAML(buf.String())
			} else {
				api.SetIngressJSON(buf.String())
			}
		case "service":
			if isYAML {
				api.SetServiceYAML(ns, name, buf.String())
			} else {
				api.SetServiceJSON(ns, name, buf.String())
			}
		case "endpoints":
			if isYAML {
				api.SetEndpointsYAML(ns, name, buf.String())
			} else {
				api.SetEndpointsJSON(ns, name, buf.String())
			}
		case "":
			log.Printf("couldn't determine kind in file: %s\n", p)
			return nil
		default:
			log.Printf("unsupported kind: %s, in file: %s\n", kind, p)
			return nil
		}

		return nil
	})

	if tlsCertPath == "" && tlsKeyPath == "" {
		if err := http.ListenAndServe(address, api); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err := http.ListenAndServeTLS(address, tlsCertPath, tlsKeyPath, api); err != nil {
			log.Fatalln(err)
		}
	}
}
