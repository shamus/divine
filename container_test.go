package divine_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shamus/divine"
)

func TestContainer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Container Suite")
}

type (
	Pizza struct {
		Toppings string
	}

	Sliced interface {
		Slices() int8
	}

	Shared interface {
		Share() bool
	}
)

func (p Pizza) Slices() int8 {
	return 8
}

var _ = Describe("Book", func() {
	var container divine.Container

	BeforeEach(func() {
		container = divine.New()
	})

	Describe("registering a dependency", func() {
		Context("by default", func() {
			var err error

			BeforeEach(func() {
				err = container.Provide(Pizza{})
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("which does not implement the specified type", func() {
			var err error

			BeforeEach(func() {
				err = container.Provide(Pizza{}, divine.AsType((*Shared)(nil)))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(fmt.Errorf("divine_test.Pizza does not implement divine_test.Shared")))
			})
		})
	})

	Describe("requesting a dependency", func() {
		Context("by default", func() {
			var pizza Pizza

			BeforeEach(func() {
				pizza = Pizza{}
				container.Provide(pizza)
			})

			It("yields the requested dependency to the dependent function", func() {
				var dependent func(p Pizza)
				Expect(container).To(Provide(pizza).To(&dependent))
			})
		})

		Context("which is unknown", func() {
			var executed bool
			var err error

			BeforeEach(func() {
				err = divine.Inject(container, func(p Pizza) {
					executed = true
				})
			})

			It("does not execute the dependent function", func() {
				Expect(executed).To(BeFalse())
			})

			It("returns an error", func() {
				Expect(err.Error()).To(BeEquivalentTo("could not fulfill request for divine_test.Pizza: You can't always get what you want (divine_test.Pizza)"))
			})
		})

		Describe("which was registered with a specific type", func() {
			var pizza Pizza

			BeforeEach(func() {
				pizza = Pizza{}
				container.Provide(pizza, divine.AsType((*Sliced)(nil)))
			})

			Context("by default", func() {
				It("yields the requested dependency to the dependent function", func() {
					var dependent func(s Sliced)
					Expect(container).To(Provide(pizza).To(&dependent))
				})
			})

			Context("when requesting by the concrete type", func() {
				var executed bool
				var err error

				BeforeEach(func() {
					err = divine.Inject(container, func(p Pizza) {
						executed = true
					})
				})

				It("does not execute the dependent function", func() {
					Expect(executed).To(BeFalse())
				})

				It("does returns an error", func() {
					Expect(err.Error()).To(BeEquivalentTo("could not fulfill request for divine_test.Pizza: You can't always get what you want (divine_test.Pizza)"))
				})
			})
		})

		Describe("which is instantiated lazily", func() {
			var mushrooms = "mushrooms"
			var mushroomPizza Pizza = Pizza{Toppings: mushrooms}

			Context("by default", func() {
				var mushroomPizzaCount = 0
				var mushroomPizzaCreator = func() Pizza {
					mushroomPizzaCount++
					return mushroomPizza
				}

				BeforeEach(func() {
					mushroomPizzaCount = 0
					container.ProvideLazily(mushroomPizzaCreator)
				})

				It("invokes the factory function to create the depenency", func() {
					var pizzaEater func(pizza Pizza)
					Expect(container).To(Provide(mushroomPizza).To(&pizzaEater))
				})

				It("caches the result of the first call to the factory for subsequent requests", func() {
					divine.Inject(container, func(pizza Pizza) {})
					divine.Inject(container, func(pizza Pizza) {})

					Expect(mushroomPizzaCount).To(Equal(1))
				})
			})

			Context("when the factory requires a dependency", func() {
				type Topping string
				var topping Topping = "pepperoni"
				var pizzaCreator = func(topping Topping) Pizza {
					return Pizza{Toppings: string(topping)}
				}

				BeforeEach(func() {
					container.Provide(topping)
					container.ProvideLazily(pizzaCreator)
				})

				It("invokes the factory function with arguments provided by the container", func() {
					var pizzaEater func(pizza Pizza)
					Expect(container).To(Provide(Pizza{Toppings: string(topping)}).To(&pizzaEater))
				})
			})

			Context("when the factory requires an unknown dependency", func() {
				type Topping string
				var pizzaCreator = func(topping Topping) Pizza {
					return Pizza{Toppings: string(topping)}
				}

				BeforeEach(func() {
					container.ProvideLazily(pizzaCreator)
				})

				It("returns an error", func() {
					err := divine.Inject(container, func(p Pizza) {})
					Expect(err.Error()).To(BeEquivalentTo("could not fulfill request for divine_test.Pizza: You can't always get what you want (divine_test.Topping)"))
				})
			})
		})

		Describe("requesting a dependency with a circular dependency", func() {
			var pizzaCreator = func(p Pizza) Pizza {
				return Pizza{Toppings: p.Toppings}
			}

			BeforeEach(func() {
				container.ProvideLazily(pizzaCreator)
			})

			It("returns an error", func() {
				err := divine.Inject(container, func(p Pizza) {})
				Expect(err.Error()).To(BeEquivalentTo("could not fulfill request for divine_test.Pizza: Circular dependency for divine_test.Pizza: it requires divine_test.Pizza which in turn requires divine_test.Pizza"))
			})
		})
	})

	Describe("when wrapping another container", func() {
		var wrappingContainer divine.Container
		var mushroomPizza Pizza = Pizza{Toppings: "mushroom"}
		var pepperoniPizza Pizza = Pizza{Toppings: "pepperoni"}

		BeforeEach(func() {
			wrappingContainer = divine.Wrap(container)
			container.Provide(mushroomPizza, divine.AsType((*Sliced)(nil)))
			wrappingContainer.Provide(pepperoniPizza)
		})

		It("returns dependencies from the current container", func() {
			var pizzaEater func(p Pizza)
			Expect(wrappingContainer).To(Provide(pepperoniPizza).To(&pizzaEater))
		})

		It("returns dependencies from the wrapped container", func() {
			var slicedPizzaEater func(s Sliced)
			Expect(wrappingContainer).To(Provide(mushroomPizza).To(&slicedPizzaEater))
		})
	})
})
