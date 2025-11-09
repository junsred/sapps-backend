package app

import (
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSetupHTTP(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	providers := httpAppConstructors()
	for d := range providers {
		provider := providers[d]
		typeProvider := reflect.TypeOf(provider)
		providerFunction := reflect.MakeFunc(typeProvider, func(args []reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.Zero(typeProvider.Out(0))}
		})
		err := c.Provide(providerFunction.Interface())
		if err != nil {
			t.Fatal(err)
		}
	}
	b := BackendApp{
		App: fiber.New(),
	}
	b.setupDigWithoutAuthHTTPRoutes()
	b.setupDigHTTPRoutes()
}
