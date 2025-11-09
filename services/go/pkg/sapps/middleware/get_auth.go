package middleware

import (
	"sapps/lib/util"

	"go.uber.org/dig"
)

type GetAuthMiddleware struct {
	dig.In
}

func (r *GetAuthMiddleware) Handler(c *RequestContext) error {
	claims, err := util.VerifyToken(c.Token())
	if err != nil {
		c.LogErr(err)
		return c.Error(StatusUnauthorized, "invalid token")
	}
	userID, err := claims.GetIssuer()
	if err != nil {
		c.LogErr(err)
		return c.Error(StatusUnauthorized, "invalid token")
	}
	c.SetUserID(userID)
	session, _ := claims["session"].(string)
	c.SetSession(session)
	return c.Next()
}
