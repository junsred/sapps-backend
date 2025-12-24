package route

import (
	"log"
	"sapps/lib/connection"
	"sapps/lib/util"
	"sapps/pkg/sapps/middleware"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/dig"
)

type PostLoginFirebase struct {
	dig.In
	PostgresMainDB *connection.PostgresMainDB
	FirebaseApp    *connection.FirebaseApp
}

func (r *PostLoginFirebase) Handler(c *middleware.RequestContext) error {
	type Request struct {
		Token    string `json:"token"`
		DeviceID string `json:"device_id"`
	}
	type Response struct {
		Token   string `json:"token"`
		NewUser bool   `json:"new_user"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusBadRequest, err.Error())
	}

	ctx := c.UserContext()
	authUser, err := r.FirebaseApp.AuthToken(ctx, req.Token)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusBadRequest, "Invalid token")
	}
	var userID string
	var newUser bool
	err = r.PostgresMainDB.QueryRow(ctx, `
	SELECT id FROM users WHERE firebase_id = $1
	`, authUser.UID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			userID = uuid.New().String()
			newUser = true
		} else {
			c.LogErr(err)
			return c.Error(middleware.StatusInternalServerError, err.Error())
		}
	}
	log.Println("firebase_id", authUser.UID, "user_id", userID, "device_id", req.DeviceID)
	session := uuid.New().String()
	token, err := util.GenerateToken(jwt.MapClaims{
		"iss":     userID,
		"iat":     time.Now().Unix(),
		"session": session,
	})
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, err.Error())
	}
	language := c.Language()
	buildNumber := c.BuildNumber()
	store := c.Store()
	insertUser := func() error {
		if newUser {
			_, err = r.PostgresMainDB.Exec(ctx,
				`INSERT INTO users (id, firebase_id, last_login, session, last_token, device_id, language, build_number, store, ip_address, country) VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8, $9, $10)`,
				userID, authUser.UID, session, token, req.DeviceID, language, buildNumber, store, c.Get("CF-Connecting-IP"), c.Get("CF-IPCountry"))
			/*if err == nil && c.BuildNumber() != nil && *c.BuildNumber() >= 8 && c.Store() != nil && (*c.Store() == "play_store" || *c.Store() == "local_source") {
				genUUID := uuid.New().String()
				// insert to premium_data
				_, err = r.PostgresMainDB.Exec(ctx,
					`INSERT INTO premium_data (id, premium_type, expire_date) VALUES ($1, $2, NOW()+INTERVAL '3 days')`,
					genUUID, "test_6m")
				if err != nil {
					log.Println("Error inserting premium_data:", err)
				}
				_, err = r.PostgresMainDB.Exec(ctx,
					`UPDATE users SET premium_id = $1 WHERE id = $2`,
					genUUID, userID)
				if err != nil {
					log.Println("Error updating users with premium data:", err)
				}
			}*/
		} else {
			_, err = r.PostgresMainDB.Exec(ctx,
				`UPDATE users SET last_login = NOW(), session = $1, last_token = $2, device_id = $3 WHERE id = $4`,
				session, token, req.DeviceID, userID)
		}
		return err
	}
	err = insertUser()
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, err.Error())
	}

	return c.JSON(Response{
		Token:   token,
		NewUser: newUser,
	})
}
