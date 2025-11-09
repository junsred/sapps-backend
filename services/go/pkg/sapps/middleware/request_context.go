package middleware

import (
	"sapps/pkg/sapps/model"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handle interface {
	Handler(c *RequestContext) error
}

func HandleWrapper(handler Handle) fiber.Handler {
	return func(c *fiber.Ctx) error {
		middlewareUserContext := NewRequestContext(c)
		return handler.Handler(middlewareUserContext)
	}
}

type RequestHandler func(*RequestContext) error

type RequestContext struct {
	*fiber.Ctx
}

func NewRequestContext(c *fiber.Ctx) *RequestContext {
	return &RequestContext{c}
}

func (c *RequestContext) Token() string {
	return c.Get("token")
}

func (c *RequestContext) UserID() string {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

func (c *RequestContext) Language() *string {
	language := c.Get("language")
	if language == "" {
		return nil
	}
	return &language
}

func (c *RequestContext) Store() *string {
	store := c.Get("store")
	if store == "" {
		return nil
	}
	return &store
}

func (c *RequestContext) BuildNumber() *int {
	buildNumber := c.Get("version")
	if buildNumber == "" {
		return nil
	}
	buildNumberInt, err := strconv.Atoi(buildNumber)
	if err != nil {
		return nil
	}
	return &buildNumberInt
}

func (c *RequestContext) SetUserID(userID string) {
	_ = c.Locals("user_id", userID)
}

func (c *RequestContext) SetSession(session string) {
	_ = c.Locals("session", session)
}

func (c *RequestContext) Session() string {
	session, ok := c.Locals("session").(string)
	if !ok {
		return ""
	}
	return session
}

func (c *RequestContext) User() *model.User {
	user, ok := c.Locals("user").(*model.User)
	if !ok {
		return nil
	}
	return user
}

func (c *RequestContext) SetUser(user *model.User) {
	_ = c.Locals("user", user)
}
