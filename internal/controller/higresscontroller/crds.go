package higresscontroller

import (
	"fmt"
	"io"
	"os"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	crdsPath          = "./config/crd/higress"
	decoderBufferSize = 4096
)

func getCRDsManifests() ([]string, error) {
	files, err := os.ReadDir(crdsPath)
	if err != nil {
		return nil, err
	}

	var manifests []string
	for _, f := range files {
		manifests = append(manifests, fmt.Sprintf("%v/%v", crdsPath, f.Name()))
	}

	return manifests, nil
}

func getCRDs() ([]*v1.CustomResourceDefinition, error) {
	manifests, err := getCRDsManifests()
	if err != nil {
		return nil, err
	}

	var crds []*v1.CustomResourceDefinition

	resolveManifest := func(path string) (err error) {
		var f *os.File
		f, err = os.Open(path)
		defer func() {
			if _err_ := f.Close(); _err_ != nil {
				err = _err_
			}
		}()

		if err != nil {
			return fmt.Errorf("failed to open the CRD manifest %v: %w", path, err)
		}

		dec := yaml.NewYAMLOrJSONDecoder(f, decoderBufferSize)
		for {
			var crd v1.CustomResourceDefinition
			if err := dec.Decode(&crd); err != nil {
				if err != io.EOF {
					return fmt.Errorf("failed to parse the CRD manifest %v: %w", path, err)
				}
				break
			}
			crds = append(crds, &crd)
		}
		return
	}

	for _, path := range manifests {
		if err = resolveManifest(path); err != nil {
			return nil, err
		}
	}

	return crds, nil
}
