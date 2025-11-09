package middleware

import (
	"context"
	"sapps/lib/connection"
	"sapps/pkg/sapps/model"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/dig"
)

type VerifyAuthMiddleware struct {
	dig.In
	PostgresMainDB *connection.PostgresMainDB
}

func (r *VerifyAuthMiddleware) Handler(c *RequestContext) error {
	userID := c.UserID()
	if userID == "" {
		return c.Error(StatusUnauthorized, "unauthorized")
	}
	ctx := c.UserContext()

	// Then, select the user's premium data and other information
	var user model.User
	var premiumExpireDate *time.Time
	var coinResetDate *time.Time
	err := r.PostgresMainDB.QueryRow(ctx, `SELECT pd.premium_type, pd.expire_date, u.firebase_token, coalesce(u.coin, 0), u.coin_reset_date, u.debug, u.special_offer_deadline
FROM users u
LEFT JOIN premium_data pd ON pd.id = u.premium_id AND (pd.expire_date is null OR pd.expire_date > NOW())
WHERE u.id = $1 AND u.session = $2`, userID, c.Session()).Scan(
		&user.PremiumType, &premiumExpireDate, &user.FirebaseToken, &user.Coin, &coinResetDate, &user.Debug, &user.SpecialOfferDeadline,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Error(StatusUnauthorized, "unauthorized")
		}
		c.LogErr(err)
		return c.Error(StatusInternalServerError, err.Error())
	}
	if user.SpecialOfferDeadline != nil && user.SpecialOfferDeadline.Before(time.Now()) {
		user.SpecialOfferDeadline = nil
	}
	language := c.Language()
	buildNumber := c.BuildNumber()
	store := c.Store()
	run := func(id string, ip string, country string, ctx context.Context) {
		_, err := r.PostgresMainDB.Exec(ctx, `UPDATE users 
SET last_online = NOW(), 
coin = case
WHEN coin_reset_date is null or coin_reset_date < NOW() THEN 2
ELSE coin
END,
coin_reset_date = case
WHEN coin_reset_date is null or coin_reset_date < NOW() THEN NOW() + INTERVAL '7 day'
ELSE coin_reset_date
END,
language = case when $2::text is null then language else $2 end,
ip_address = case when $3::text is null or $3::text = '' then ip_address else $3 end,
country = case when $4::text is null or $4::text = '' or $4::text = 'XX' or $4::text = 'T1' then country else $4 end,
build_number = case when $5::integer is null then build_number else $5 end,
store = case when $6::text is null then store else $6 end
WHERE id = $1`, id, language, ip, country, buildNumber, store)
		if err != nil {
			c.LogErr(err)
		}
	}
	if coinResetDate == nil || time.Now().After(*coinResetDate) {
		run(userID, c.Get("CF-Connecting-IP"), c.Get("CF-IPCountry"), ctx)
		user.Coin = 2
	} else {
		go run(userID, c.Get("CF-Connecting-IP"), c.Get("CF-IPCountry"), context.Background())
	}

	user.ID = userID
	if premiumExpireDate != nil {
		user.PremiumExpireDate = premiumExpireDate.Unix()
	}
	c.SetUser(&user)
	return c.Next()
}
