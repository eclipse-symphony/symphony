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
		instance  model.InstanceState
	}
	encodable interface{ UnstructuredContent() map[string]interface{} }
)

var (
	decoder = serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	_ postrender.PostRenderer = &PostRenderer{}
)

// Run implements PostRenderer.
func (r *PostRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
    modifiedManifests := new(bytes.Buffer)
    reader := yaml.NewYAMLReader(bufio.NewReader(renderedManifests))

    for {
        manifest, err := reader.Read()
        if err != nil {
            if errors.Is(err, io.EOF) {
                break
            }
            return nil, err
        }

        // Trim spaces/newlines. Comment-only docs still have non-empty bytes but
        // they're not actual k8s objects. We handle that by trying to decode first.
        if len(bytes.TrimSpace(manifest)) == 0 {
            // completely empty doc -> skip
            continue
        }

        // Try to decode into an Unstructured object
        obj, _, decErr := decoder.Decode(manifest, nil, nil)
        if decErr != nil {
            // This happens for docs that are only comments or otherwise not K8s
            // resources (like the PodDisruptionBudget comment block).
            // We just ignore those docs instead of failing the whole render.
            continue
        }

        // Inject Symphony metadata
        if err := r.populateMeta(obj); err != nil {
            return nil, err
        }

        // Only now that we know it's a valid resource do we write it out.
        // If you still want the "# Source: ..." line, keep getSourceBytes().
        // Otherwise you can drop it for cleanliness.
        modifiedManifests.WriteString("---\n")
        modifiedManifests.Write(r.getSourceBytes(manifest))

        if err := r.encodeInto(modifiedManifests, obj); err != nil {
            return nil, err
        }
    }

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
