package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	ectx "github.com/eclipse-symphony/symphony/packages/testutils/internal/context"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/google/uuid"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type (
	ListRunner interface {
		Run() ([]*release.Release, error)
	}
	HelmExpectation struct {
		// pattern is the pattern of the release
		pattern string

		// description is a friendly description of the expectation
		description string

		// removed indicates whether the release is expected to be present or not
		removed bool

		// namespace is the namespace of the release
		namespace string

		// releaseCondition specifies the types.that the release should satisfy
		releaseCondition     types.Condition
		releaseListCondition types.Condition

		action        ListRunner
		actionBuilder func() (ListRunner, error)
		l             func(format string, args ...interface{})
		tick          time.Duration
		timeout       time.Duration
		nameRegex     *regexp.Regexp
		level         int
		id            string
		initialised   bool
	}

	Option func(*HelmExpectation)
)

const (
	defaultTick    = 5 * time.Second
	defaultTimeout = 5 * time.Minute
)

var _ types.Expectation = &HelmExpectation{}

// NewExpectation creates a new helm expectation.
func NewExpectation(pattern, namespace string, opts ...Option) (*HelmExpectation, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}

	he := &HelmExpectation{
		pattern:   boundPattern(pattern),
		namespace: namespace,
		tick:      defaultTick,
		timeout:   defaultTimeout,
		id:        uuid.NewString(),
	}
	he.actionBuilder = getDefaultActionBuilder(namespace, he.log)

	for _, opt := range opts {
		opt(he)
	}

	nameRegex, err := regexp.Compile(he.pattern)
	if err != nil {
		return nil, err
	}
	he.nameRegex = nameRegex

	he.initializeCountCondition()

	return he, nil
}

// MustNew creates a new helm expectation. It panics if the expectation cannot be created.
func MustNew(name, namespace string, opts ...Option) *HelmExpectation {
	he, err := NewExpectation(name, namespace, opts...)
	if err != nil {
		panic(err)
	}
	return he
}

// NewPresent creates a new helm expectation that expects the release to be present.
func MustNewAbsent(name, namespace string, opts ...Option) *HelmExpectation {
	return MustNew(name, namespace, append(opts, WithRemoved(true))...)
}

func (he *HelmExpectation) initAction() error {
	if he.initialised {
		return nil
	}
	action, err := he.actionBuilder()
	if err != nil {
		return err
	}
	he.action = action
	he.initialised = true
	return nil
}

func (he *HelmExpectation) log(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	format = "%s[%s]: %s\n"
	args = []interface{}{strings.Repeat(" ", he.level), he.Description(), s}

	if he.l != nil {
		he.l(format, args...)
	} else {
		logger.GetDefaultLogger()(format, args...)
	}
}

// Verify implements types.Expectation.
func (he *HelmExpectation) Verify(c context.Context) error {
	ctx := ectx.From(c)
	he.level = ctx.Level()

	return helpers.Eventually(ctx, func(ctx context.Context) (err error) {
		he.log(strings.Repeat("-", 80))
		he.log("Verifying helm release")
		defer func() {
			if err != nil {
				he.log("Error while verifying helm release: %s", err)
			}
		}()

		matches, err := he.getResults(ctx)
		if err != nil {
			return
		}
		if err = he.verifyConditions(ctx, matches); err != nil {
			return
		}
		return nil
	}, he.tick, "Timed out while verifying helm release %s", he.Description())
}

// Id implements types.Expectation.
func (he *HelmExpectation) Id() string {
	return he.id
}

func (re *HelmExpectation) initializeCountCondition() {
	countCondition := conditions.GreaterThan(0)
	if re.removed {
		countCondition = conditions.Count(0)
	}

	addCondition(&re.releaseListCondition, countCondition)
}

func (he *HelmExpectation) getResults(ctx context.Context) ([]*release.Release, error) {
	if err := he.initAction(); err != nil {
		return nil, err
	}
	releases, err := he.action.Run()
	if err != nil {
		return nil, err
	}
	he.log("Found %d releases", len(releases))
	he.log("Action Type: %T", he.action)
	matches := he.getMatches(releases)
	he.log("Found %d matching releases", len(matches))
	return matches, nil
}

func (he *HelmExpectation) verifyConditions(ctx context.Context, releases []*release.Release) error {
	//Todo: Add support for list conditions
	he.log("Verifying conditions")
	arr := make([]map[string]interface{}, len(releases))
	b, err := json.Marshal(releases)
	if err != nil {
		return err
	}
	json.Unmarshal(b, &arr)

	if err := he.verifyListCondition(ctx, arr); err != nil {
		return err
	}

	if err := he.verifyUnitCondition(ctx, arr); err != nil {
		return err
	}

	return nil
}

func (he *HelmExpectation) verifyUnitCondition(c context.Context, releases []map[string]interface{}) error {
	ctx := ectx.From(c)
	if he.releaseCondition != nil {
		for _, release := range releases {
			if err := he.releaseCondition.IsSatisfiedBy(ctx.Nested(), release); err != nil {
				return err
			}
		}
	}
	return nil
}

func (he *HelmExpectation) verifyListCondition(c context.Context, releases []map[string]interface{}) error {
	ctx := ectx.From(c)
	if he.releaseListCondition != nil {
		return he.releaseListCondition.IsSatisfiedBy(ctx.Nested(), releases)
	}
	return nil
}

func (he *HelmExpectation) getMatches(releases []*release.Release) (matches []*release.Release) {
	for i := range releases {
		if he.nameRegex.MatchString(releases[i].Name) {
			matches = append(matches, releases[i])
		}
	}
	return matches
}

func (he *HelmExpectation) Description() string {
	if he.description != "" {
		return he.description
	}
	return fmt.Sprintf("helm expectation: %s", he.pattern)
}

func getDefaultActionBuilder(namespace string, logger logger.Logger) func() (ListRunner, error) {
	return func() (ListRunner, error) {
		settings := cli.New()
		var allNamespaces bool

		if namespace == "*" {
			allNamespaces = true
		} else {
			settings.SetNamespace(namespace)
		}

		actionConfig := new(action.Configuration)
		actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), logger)
		action := action.NewList(actionConfig)

		action.AllNamespaces = allNamespaces
		return action, nil
	}
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

func createValueConditionFrom(condition types.Condition, isListCondition bool) types.Condition {
	return conditions.Basic(func(ctx context.Context, i interface{}) error {
		if isListCondition {
			switch releases := i.(type) {
			case []map[string]any:
				values := make([]map[string]interface{}, len(releases))
				for i, release := range releases {
					val, _ := release["config"].(map[string]interface{})
					values[i] = val
				}
				return condition.IsSatisfiedBy(ctx, values)
			default:
				return fmt.Errorf("expected []interface{}, got %T", i)
			}
		}
		switch release := i.(type) {
		case map[string]interface{}:
			value, _ := release["config"].(map[string]interface{})
			return condition.IsSatisfiedBy(ctx, value)

		default:
			return fmt.Errorf("expected map[string]interface{}, got %T", i)
		}
	}, conditions.WithBasicDescription("values check"))
}

func addCondition(existingCondition *types.Condition, newCondition types.Condition) {
	if *existingCondition != nil {
		if c, ok := (*existingCondition).(interface {
			And(...types.Condition) types.Condition
		}); ok {
			*existingCondition = c.And(newCondition)
		}
	} else {
		*existingCondition = conditions.All(newCondition)
	}
}
