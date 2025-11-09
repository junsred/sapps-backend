package route

import (
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"

	"go.uber.org/dig"
)

type PatchAccount struct {
	dig.In
	MainDB *maindb.MainDB
}

func (r *PatchAccount) Handler(c *middleware.RequestContext) error {
	type Request struct {
		FirebaseToken          *string                `json:"firebase_token"`
		NotificationPermission *bool                  `json:"notification_permission"`
		Timezone               *string                `json:"timezone"`
		DeviceInfo             map[string]interface{} `json:"device_info"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusBadRequest, err.Error())
	}
	user := c.User()
	if req.FirebaseToken != nil {
		var usingID *string
		tx, err := r.MainDB.Begin(c.UserContext())
		if err != nil {
			c.LogErr(err)
			return c.Error(middleware.StatusInternalServerError, err.Error())
		}
		defer tx.Rollback(c.UserContext())
		err = tx.QueryRow(c.UserContext(), `
		SELECT id FROM users WHERE firebase_token = $1
		`, req.FirebaseToken).Scan(&user.ID)
		if err == nil {
			_, err = tx.Exec(c.UserContext(), `
			UPDATE users
			SET firebase_token = $1
			WHERE id = $2
			`, nil, usingID)
			if err != nil {
				c.LogErr(err)
			}
		}
		_, err = tx.Exec(c.UserContext(), `
		UPDATE users
		SET firebase_token = $1
		WHERE id = $2
		`, req.FirebaseToken, user.ID)
		if err != nil {
			c.LogErr(err)
		}
		err = tx.Commit(c.UserContext())
		if err != nil {
			c.LogErr(err)
		}
	}
	if req.NotificationPermission != nil {
		_, err := r.MainDB.Exec(c.UserContext(), `
		UPDATE users
		SET notification_permission = $1
		WHERE id = $2
		`, req.NotificationPermission, user.ID)
		if err != nil {
			c.LogErr(err)
		}
	}
	if req.Timezone != nil {
		_, err := r.MainDB.Exec(c.UserContext(), `
		UPDATE users
		SET timezone = $1
		WHERE id = $2
		`, req.Timezone, user.ID)
		if err != nil {
			c.LogErr(err)
		}
	}
	if req.DeviceInfo != nil {
		_, err := r.MainDB.Exec(c.UserContext(), `
		UPDATE users
		SET device_info = $1
		WHERE id = $2
		`, req.DeviceInfo, user.ID)
		if err != nil {
			c.LogErr(err)
		}
	}
	return c.JSON(user)
}
