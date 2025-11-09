package route

import (
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"sapps/pkg/sapps/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type PostRevenuecatWebhook struct {
	dig.In
	MainDB *maindb.MainDB
}

func (r *PostRevenuecatWebhook) Handler(c *middleware.RequestContext) error {
	var eventData service.RevenueCatEvent
	if err := c.BodyParser(&eventData); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusBadRequest, err.Error())
	}

	revenueCatService := service.NewRevenueCatService(r.MainDB)
	if err := revenueCatService.HandleWebhook(c.Context(), &eventData); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Webhook processed successfully",
	})
}
