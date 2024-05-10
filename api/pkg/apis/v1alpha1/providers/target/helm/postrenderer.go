package helm

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils/metahelper"
	"helm.sh/helm/v3/pkg/postrender"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
	syaml "sigs.k8s.io/yaml"
)

type (
	PostRenderer struct {
		populator metahelper.MetaPopulator
		instance  model.InstanceSpec
	}
	encodable interface{ UnstructuredContent() map[string]interface{} }
)

var (
	decoder = serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	_ postrender.PostRenderer = &PostRenderer{}
)

// Run implements PostRenderer.
func (r *PostRenderer) Run(renderedManifests *bytes.Buffer) (modifiedManifests *bytes.Buffer, err error) {
	modifiedManifests = new(bytes.Buffer)
	reader := yaml.NewYAMLReader(bufio.NewReader(renderedManifests))

	for {
		// Read the next YAML document
		manifest, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		// Skip empty manifests
		if len(manifest) == 0 {
			continue
		}
		// Source is the first line of the manifest.
		modifiedManifests.WriteString("---\n")
		modifiedManifests.Write(r.getSourceBytes(manifest))

		// Decode the manifest into a Kubernetes resource
		obj, _, err := decoder.Decode(manifest, nil, nil)
		if err != nil {
			return nil, err
		}

		// Apply the metadata to the resource
		if err := r.populateMeta(obj); err != nil {
			return nil, err
		}

		// Re-encode the manifest and write it to the modifiedManifests buffer
		// We won't use the codec because it encodes in JSON, not YAML
		if err := r.encodeInto(modifiedManifests, obj); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
	}
	syaml.Marshal(modifiedManifests)

	return modifiedManifests, nil
}

func (r *PostRenderer) populateMeta(obj runtime.Object) error {
	switch obj := obj.(type) {
	case *unstructured.Unstructured:
		if err := r.populator.PopulateMeta(obj, r.instance); err != nil {
			return err
		}
	case *unstructured.UnstructuredList:
		for _, item := range obj.Items {
			if err := r.populator.PopulateMeta(&item, r.instance); err != nil {
				return err
			}
		}
	default:
		return errors.New("manifest is not a Kubernetes resource")
	}
	return nil
}

func (r *PostRenderer) encodeInto(buffer io.Writer, obj runtime.Object) error {
	switch obj := obj.(type) {
	case *unstructured.Unstructured:
		if err := encodeInto(buffer, obj); err != nil {
			return err
		}
	case *unstructured.UnstructuredList:
		if err := encodeInto(buffer, obj); err != nil {
			return err
		}
	default:
		return errors.New("manifest is not a Kubernetes resource")
	}
	return nil
}

func encodeInto(writer io.Writer, obj encodable) (err error) {
	bytes, err := syaml.Marshal(obj.UnstructuredContent())
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	return err
}

func (r *PostRenderer) getSourceBytes(manifest []byte) []byte {
	// the very first manifest usually starts with "---\n"
	// we need to remove it to get the source
	var startIndex int
	if bytes.HasPrefix(manifest, []byte("---\n")) {
		startIndex = 4
	}
	return manifest[startIndex : bytes.IndexByte(manifest[startIndex:], '\n')+1+startIndex]
}
