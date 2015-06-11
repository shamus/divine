package divine

import (
	"fmt"
	"reflect"
)

type (
	Dependent interface{}

	Container interface {
		Provide(item interface{}, as ...reflect.Type) error
		ProvideLazily(factory Dependent)
		ByType(need reflect.Type) (interface{}, error)
	}

	InjectError struct {
		Requested reflect.Type
		Err       error
	}

	simple struct {
		static    map[reflect.Type]interface{}
		factories map[reflect.Type]interface{}
	}

	wrapped struct {
		Container
		parent Container
	}
)

func isCircular(need reflect.Type, signature reflect.Type, factories map[reflect.Type]interface{}) bool {
	for i := 0; i < signature.NumIn(); i++ {
		param := signature.In(i)
		factory, found := factories[param]
		if !found {
			continue
		}

		dependentSignature := reflect.TypeOf(factory)
		for j := 0; j < dependentSignature.NumIn(); j++ {
			dependentParam := dependentSignature.In(j)
			if dependentParam == need {
				return true
			}
		}
	}

	return false
}

func AsType(i interface{}) reflect.Type {
	return reflect.TypeOf(i).Elem()
}

func Inject(c Container, d Dependent) error {
	signature := reflect.TypeOf(d)
	provided := make([]reflect.Value, signature.NumIn())

	for i := 0; i < signature.NumIn(); i++ {
		param := signature.In(i)
		p, err := c.ByType(param)
		if err != nil {
			return InjectError{Requested: param, Err: err}
		}

		provided[i] = reflect.ValueOf(p)
	}

	reflect.ValueOf(d).Call(provided)
	return nil
}

func MustInject(c Container, d Dependent) {
	if err := Inject(c, d); err != nil {
		panic(err)
	}
}

func New() Container {
	static := make(map[reflect.Type]interface{})
	factories := make(map[reflect.Type]interface{})

	return &simple{static, factories}
}

func Wrap(c Container) Container {
	return &wrapped{Container: New(), parent: c}
}

func (c *simple) Provide(item interface{}, as ...reflect.Type) error {
	need := reflect.TypeOf(item)
	if len(as) == 0 {
		c.static[need] = item
		return nil
	}

	var err error
	for _, t := range as {
		if !(need.Implements(t)) {
			err = fmt.Errorf("%v does not implement %v", need, t)
			break
		}

		c.static[t] = item
	}

	return err
}

func (c *simple) ProvideLazily(factory Dependent) {
	need := reflect.TypeOf(factory).Out(0)
	c.factories[need] = factory
}

func (c *simple) ByType(need reflect.Type) (interface{}, error) {
	if provided, found := c.static[need]; found {
		return provided, nil
	}

	factory, found := c.factories[need]
	if !found {
		return nil, fmt.Errorf("You can't always get what you want (%v)", need)
	}

	signature := reflect.TypeOf(factory)
	provided := make([]reflect.Value, signature.NumIn())

	for i := 0; i < signature.NumIn(); i++ {
		param := signature.In(i)
		if isCircular(need, signature, c.factories) {
			return nil, fmt.Errorf("Circular dependency for %v: it requires %v which in turn requires %v", need, param, need)
		}

		p, err := c.ByType(param)
		if err != nil {
			return nil, err
		}

		provided[i] = reflect.ValueOf(p)
	}

	factoried := reflect.ValueOf(factory).Call(provided)[0].Interface()
	c.static[need] = factoried

	return factoried, nil
}

func (c *wrapped) ByType(need reflect.Type) (interface{}, error) {
	value, err := c.Container.ByType(need)
	if err != nil {
		value, err = c.parent.ByType(need)
	}

	return value, err
}

func (ie InjectError) Error() string {
	return fmt.Sprintf("could not fulfill request for %v: %v", ie.Requested, ie.Err.Error())
}
