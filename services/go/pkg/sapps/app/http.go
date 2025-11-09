package app

import (
	"fmt"
	"sapps/pkg/sapps/middleware"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/dig"
)

type BackendApp struct {
	*fiber.App
	HttpClient *http.Client
}

var c = dig.New()

func mustInvoke[T any]() *T {
	retT := new(T)
	err := c.Invoke(func(l T) {
		retT = &l
	})
	if err != nil {
		panic(fmt.Sprintf("%v, %v", reflect.TypeOf(retT), err))
	}

	return retT
}

func (b *BackendApp) setupHTTPWebRoutes() {
	for _, serviceConstructor := range httpAppConstructors() {
		if err := c.Provide(serviceConstructor); err != nil {
			panic(err)
		}
	}
	b.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins:     "*",
		AllowCredentials: false,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))
	b.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			debugStack := string(debug.Stack())
			reqC := middleware.NewRequestContext(c)
			reqC.LogErr(fmt.Errorf("%v\n%v", e, debugStack), "critical")
			reqC.ErrorInternal("Internal Server Error")
		},
	}))
	b.Use(middleware.HandleWrapper(mustInvoke[middleware.HandleErrorMiddleware]()))
	b.setupDigWithoutAuthHTTPRoutes()
	authMiddleware := middleware.HandleWrapper(mustInvoke[middleware.GetAuthMiddleware]())
	verifyAuthMiddleware := middleware.HandleWrapper(mustInvoke[middleware.VerifyAuthMiddleware]())
	b.setupDigHTTPRoutes(authMiddleware, verifyAuthMiddleware)

	b.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})
}

func NewHTTWebPApp() *BackendApp {
	timeoutTime := 120 * time.Second
	b := &BackendApp{
		App: fiber.New(fiber.Config{ReadTimeout: timeoutTime,
			WriteTimeout:          timeoutTime,
			IdleTimeout:           timeoutTime,
			Immutable:             true,
			BodyLimit:             30 * 1024 * 1024,
			Concurrency:           1024 * 8,
			DisableStartupMessage: false,
		}),
		HttpClient: &http.Client{},
	}
	b.setupHTTPWebRoutes()
	log.Println("HTTP WEB APP SETUP")
	return b
}
