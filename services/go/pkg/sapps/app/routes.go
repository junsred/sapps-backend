package app

import (
	"sapps/pkg/sapps/middleware"
	route "sapps/pkg/sapps/route"

	"github.com/gofiber/fiber/v2"
)

func (b *BackendApp) setupDigWithoutAuthHTTPRoutes() {
	b.Post("/webhook/revenuecat/face", middleware.HandleWrapper(mustInvoke[route.PostRevenuecatWebhook]()))
	b.Post("/webhook/kie/callback", middleware.HandleWrapper(mustInvoke[route.PostGenerativeAICallback]()))

	b.Get("/cdn/img/:id", middleware.HandleWrapper(mustInvoke[route.GetCDNImage]()))
	b.Post("/login/firebase", middleware.HandleWrapper(mustInvoke[route.PostLoginFirebase]()))
}

func (b *BackendApp) setupDigHTTPRoutes(middlewares ...fiber.Handler) {
	b.Get("/users/account", append(middlewares, middleware.HandleWrapper(mustInvoke[route.GetAccount]()))...)
	b.Patch("/users/account", append(middlewares, middleware.HandleWrapper(mustInvoke[route.PatchAccount]()))...)
	b.Post("/upload-image", append(middlewares, middleware.HandleWrapper(mustInvoke[route.PostUploadImage]()))...)
	b.Post("/scans", append(middlewares, middleware.HandleWrapper(mustInvoke[route.PostScan]()))...)
	b.Get("/scans", append(middlewares, middleware.HandleWrapper(mustInvoke[route.GetScans]()))...)
	b.Get("/scans/:id", append(middlewares, middleware.HandleWrapper(mustInvoke[route.GetScan]()))...)
	b.Post("/generative-ai", append(middlewares, middleware.HandleWrapper(mustInvoke[route.PostGenerativeAI]()))...)
	b.Get("/generative-ai/:id", append(middlewares, middleware.HandleWrapper(mustInvoke[route.GetGenerativeAI]()))...)
}
