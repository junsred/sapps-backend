package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type HandleErrorMiddleware struct {
	dig.In
}

func (r *HandleErrorMiddleware) Handler(c *RequestContext) error {
	err := c.Next()
	if err != nil {
		c.LogErr(err)
	} else if c.Response().StatusCode() < 200 || c.Response().StatusCode() >= 300 {
		body := string(c.Response().Body())
		if strings.Contains(body, "Not Found") {
			return err
		}
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		log.Printf("Error in route: %s, %d, %s\n", string(c.Request().RequestURI()), c.Response().StatusCode(), body)
	}
	if string(c.Response().Header.ContentType()) == fiber.MIMEApplicationJSON {
		c.Response().Header.SetContentType(fiber.MIMEApplicationJSONCharsetUTF8)
	}
	return err
}
