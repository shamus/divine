Divine
======

A dependency injector.

Divine injects dependencies into a function by inspecting its signature and provides the arguments based on their type. If the dependency cannot be found, an error is returned and the function isn't executed.

Features:
- Statically & Lazily provided dependencies
- Circular dependency detection
- Chaining dependency lookups across containers

## Examples

**Provide a dependency by its concrete type**
```Go
package main

import (
  "fmt"
  "github.com/shamus/divine"
)

type (
	Configuration struct {
    Foo string
  }
)

func main() {
  container := divine.New()
	container.Provide(Configuration{Foo: "value"})
  divine.Inject(container, func(configuration Configuration) {
    fmt.Println(configuration.Foo)
  })
}
```

**Provide a dependency as another type**
```Go
package main

import (
  "fmt"
  "github.com/shamus/divine"
)

type (
	Configuration interface {
    Foo() string
  }

  simpleConfiguration struct{}
)

func (c simpleConfiguration) Foo() string {
	return "value"
}

func main() {
  container := divine.New()
  container.Provide(simpleConfiguration{}, divine.AsType((*Configuration)(nil)))
  divine.Inject(container, func(configuration Configuration) {
    fmt.Println(configuration.Foo())
  })
}
```

**Provide a dependency lazily**
```Go
package main

import (
  "fmt"
  "github.com/shamus/divine"
)

type (
	Configuration struct {
    Foo string
  }
)

func main() {
  container := divine.New()
	container.ProvideLazily(func() {
    return Configuration{Foo: "value"}
  })

  divine.Inject(container, func(configuration Configuration) {
    fmt.Println(configuration.Foo)
  })
}
```
