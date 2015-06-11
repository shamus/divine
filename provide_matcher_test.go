package divine_test

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
	"github.com/shamus/divine"
)

type (
	providerMatcher struct {
		executed                     bool
		err                          error
		dependent, required, yielded interface{}
	}
)

func Provide(required interface{}) *providerMatcher {
	return &providerMatcher{required: required}
}

func (p *providerMatcher) To(dependent interface{}) types.GomegaMatcher {
	p.dependent = dependent
	return p
}

func (p *providerMatcher) Match(actual interface{}) (success bool, err error) {
	dependent := reflect.ValueOf(p.dependent).Elem()
	v := reflect.MakeFunc(dependent.Type(), func(in []reflect.Value) []reflect.Value {
		p.executed = true
		p.yielded = in[0].Interface()

		return []reflect.Value{}
	})
	dependent.Set(v)

	p.err = divine.Inject(actual.(divine.Container), dependent.Interface())
	return p.err == nil && p.executed && p.yielded == p.required, nil
}

func (p *providerMatcher) FailureMessage(actual interface{}) (message string) {
	if p.err != nil {
		return fmt.Sprintf("Expected container to yield %v, but received an error %v", p.required, p.err)
	}

	return fmt.Sprintf("Expected container to yield %v, but instead received %v", p.required, p.yielded)
}

func (p *providerMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ""
}
