package route

import (
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type GetAccount struct {
	dig.In
	MainDB *maindb.MainDB
}

func (r *GetAccount) Handler(c *middleware.RequestContext) error {
	type Response struct {
		Offerings          fiber.Map `json:"offerings"`
		Coin               int       `json:"coin"`
		PremiumType        *string   `json:"premium_type"`
		PremiumExpireDate  int64     `json:"premium_expire_date,omitempty"`
		Debug              *bool     `json:"debug,omitempty"`
		AlternativePaywall *bool     `json:"alternative_paywall"`
	}
	user := c.User()
	var premiumProducts []string
	aorb := "a"
	borb := "a"
	if user.SpecialOfferDeadline != nil {
		aorb = "b"
		borb = "b"
	}
	premiumProducts = []string{"sappsr_pro_" + aorb + "_1w", "sappsr_pro_" + borb + "_1m"}
	if user.PremiumType != nil {
		premiumType := *user.PremiumType
		if strings.HasSuffix(premiumType, "_1w") {
			premiumProducts = []string{"sappsr_pro_" + borb + "_1m"}
		} else if strings.HasSuffix(premiumType, "_1m") {
			premiumProducts = []string{}
		}
	}
	var special fiber.Map
	if user.SpecialOfferDeadline != nil {
		special = fiber.Map{
			"discount": int(87),
			"duration": int(time.Until(*user.SpecialOfferDeadline).Seconds()),
		}
	}
	resp := Response{
		Offerings: fiber.Map{
			"premium":   premiumProducts,
			"packages":  []string{"$rc_annual", "$rc_weekly"},
			"special":   special,
			"discounts": []string{"sappsr_pro_c_1w_trial_2"},
		},
		Coin:              user.Coin,
		PremiumType:       user.PremiumType,
		PremiumExpireDate: user.PremiumExpireDate,
		Debug:             user.Debug,
	}
	boolVal := false
	resp.AlternativePaywall = &boolVal

	return c.JSON(resp)
}
