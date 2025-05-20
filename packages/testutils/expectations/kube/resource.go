package kube

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	ectx "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type (
	KubeExpectation struct {
		// pattern is the prefix of the resource pattern
		pattern string

		// description is a friendly description of the expectation
		description string

		// gvk is the group, version, kind of the resource
		gvk schema.GroupVersionKind

		// namespace is the namespace of the resource if applicable
		namespace string

		// Removed indicates whether the resource is expected to be present or not
		removed bool

		// conditions specifies the condition that the resource should satisfy
		condition types.Condition

		// listCondition specifies the condition that the list of resources should satisfy
		listCondition types.Condition

		discoveryClient        discovery.DiscoveryInterface
		dynamicClient          dynamic.Interface
		mapper                 meta.RESTMapper
		discoveryClientBuilder func() (discovery.DiscoveryInterface, error)
		dynamicClientBuilder   func() (dynamic.Interface, error)

		tick        time.Duration
		l           func(format string, args ...interface{})
		nameRegex   *regexp.Regexp
		level       int
		id          string
		initialized bool
	}

	Option func(*KubeExpectation)
)

const (
	defaultTimeout = 10 * time.Minute
	defaultTick    = 10 * time.Second
)

var (
	_ types.Expectation = &KubeExpectation{}
)

func Resource(pattern, namespace string, gvk schema.GroupVersionKind, opts ...Option) (*KubeExpectation, error) {
	re := KubeExpectation{
		pattern:                boundPattern(pattern),
		gvk:                    gvk,
		tick:                   defaultTick,
		discoveryClientBuilder: defaultDiscoveryClientBuilder,
		dynamicClientBuilder:   defaultDynamicClientBuilder,
		namespace:              namespace,
		id:                     uuid.NewString(),
	}

	compiled, err := regexp.Compile(boundPattern(pattern))
	if err != nil {
		return nil, err
	}
	re.nameRegex = compiled

	for _, opt := range opts {
		opt(&re)
	}

	re.initializeCountCondition()
	return &re, nil
}

func (re *KubeExpectation) initClients() error {
	if re.initialized {
		return nil
	}
	discoveryClient, err := re.discoveryClientBuilder()
	if err != nil {
		return err
	}
	re.discoveryClient = discoveryClient

	dynamicClient, err := re.dynamicClientBuilder()
	if err != nil {
		return err
	}
	re.dynamicClient = dynamicClient
	re.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(re.discoveryClient))
	re.initialized = true
	return nil
}

func (re *KubeExpectation) initializeCountCondition() {
	countCondition := conditions.GreaterThan(0)
	if re.removed {
		countCondition = conditions.Count(0)
	}
	if re.listCondition != nil {
		re.listCondition = conditions.All(countCondition, re.listCondition)
	} else {
		re.listCondition = countCondition
	}
}

func (re *KubeExpectation) log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	format = "%s[%s]: %s\n"
	args = []interface{}{strings.Repeat(" ", re.level), re.Description(), s}

	if re.l != nil {
		re.l(format, args...)
	} else {
		logger.GetDefaultLogger()(format, args...)
	}
}

// Verify implements types.Expectation.
func (re *KubeExpectation) Verify(c context.Context) error {
	ctx := ectx.From(c)
	re.level = ctx.Level()

	return helpers.Eventually(ctx, func(ctx context.Context) error {
		re.log(strings.Repeat("-", 80))
		re.log(`Verifying resource`)
		err := re.verify(ctx, re.condition)
		if err != nil {
			re.log("Resource verification failed: %v", err)
			return err
		}
		return nil
	}, re.tick, "Timed out while verifying resource %s of kind: [%s]", re.pattern, re.gvk.String())
}

func (re *KubeExpectation) Description() string {
	if re.description != "" {
		return re.description
	}
	return fmt.Sprintf("%s: %s", re.gvk.String(), re.pattern)
}

func (re *KubeExpectation) Id() string {
	return re.id
}

func (re *KubeExpectation) getResults(ctx context.Context) ([]*unstructured.Unstructured, error) {
	if err := re.initClients(); err != nil {
		return nil, err
	}
	var namespaced bool

	mapping, err := re.mapper.RESTMapping(re.gvk.GroupKind(), re.gvk.Version)
	if err != nil {
		return nil, err
	}
	namespaced = mapping.Scope.Name() == meta.RESTScopeNameNamespace

	if namespaced && re.namespace == "" {
		return nil, fmt.Errorf("namespace is required for namespaced resources")
	}

	var list *unstructured.UnstructuredList

	if namespaced {
		namespace := re.namespace
		if namespace == "*" {
			namespace = metav1.NamespaceAll
		}
		list, err = re.dynamicClient.Resource(mapping.Resource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		list, err = re.dynamicClient.Resource(mapping.Resource).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	return re.getMatches(list), nil
}

func (re *KubeExpectation) verify(ctx context.Context, condition types.Condition) (err error) {

	matches, err := re.getResults(ctx)
	if err != nil {
		return err
	}
	re.log("Resource matches returned. %d matches", len(matches))

	return re.verifyConditions(ctx, matches)
}

func (re *KubeExpectation) verifyConditions(ctx context.Context, matches []*unstructured.Unstructured) (err error) {
	err = re.evaluateListCondition(ctx, matches, re.listCondition)
	if err != nil {
		return
	}

	err = re.evaluateCondition(ctx, matches, re.condition)
	if err != nil {
		return
	}

	return nil
}

func (re *KubeExpectation) evaluateCondition(c context.Context, objects []*unstructured.Unstructured, condition types.Condition) (err error) {
	ctx := ectx.From(c)
	if condition != nil {
		for _, object := range objects {
			err = condition.IsSatisfiedBy(ctx.Nested(), object.Object)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (re *KubeExpectation) evaluateListCondition(c context.Context, objects []*unstructured.Unstructured, condition types.Condition) (err error) {
	ctx := ectx.From(c)
	if condition != nil {
		err = condition.IsSatisfiedBy(ctx.Nested(), objects)
		if err != nil {
			return err
		}
	}

	return nil
}

func (re *KubeExpectation) getMatches(list *unstructured.UnstructuredList) []*unstructured.Unstructured {
	matches := make([]*unstructured.Unstructured, 0)
	for i := range list.Items {
		if re.nameRegex.MatchString(list.Items[i].GetName()) {
			matches = append(matches, &list.Items[i])
		}
	}
	return matches
}

func boundPattern(str string) string {
	if !strings.HasPrefix(str, "^") {
		str = "^" + str
	}
	if !strings.HasSuffix(str, "$") {
		str = str + "$"
	}
	return str
}

func defaultDiscoveryClientBuilder() (discovery.DiscoveryInterface, error) {
	return helpers.DiscoveryClient()
}

func defaultDynamicClientBuilder() (dynamic.Interface, error) {
	return helpers.DynamicClient()
}
