package model

import "time"

type User struct {
	ID                   string
	FirebaseToken        *string
	PremiumType          *string
	PremiumExpireDate    int64
	Coin                 int
	Debug                *bool
	SpecialOfferDeadline *time.Time
}
