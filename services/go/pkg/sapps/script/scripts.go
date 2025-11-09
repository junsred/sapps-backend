package script

import (
	"context"
	"sapps/lib/connection"
	"sapps/lib/util"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/v4/messaging"
)

func Scripts() {
	//go apiIsLive()
	go SendPushToNonPremiumUsers()
}

func apiIsLive() {
	ticker := time.NewTicker(45 * time.Second)
	for range ticker.C {
		func() {
			dbConn := connection.InjectMainDB()
			ctx := context.Background()
			var result int
			err := dbConn.QueryRow(ctx, `select 1`).Scan(&result)
			if err != nil {
				util.LogErr(err)
			}
			if result == 1 {
				request, err := http.NewRequest("GET", "https://hc-ping.com/c39298db-0e0f-4f1c-99ec-79365db60105", nil)
				if err != nil {
					util.LogErr(err)
				}
				client := &http.Client{}
				client.Do(request)
			} else {
				log.Println("api is not live")
			}
		}()
	}
}

func SendPushToNonPremiumUsers() {
	firebaseApp := connection.InjectFirebase()
	dbConn := connection.InjectMainDB()
	ticker := time.NewTicker(45 * time.Second)
	for range ticker.C {
		func() {
			ctx := context.Background()

			rows, err := dbConn.Query(ctx, "SELECT firebase_token, language FROM users WHERE premium_id IS NULL AND firebase_token IS NOT NULL and firebase_token != '' AND now() >= registered_at + INTERVAL '15 minutes' and register_not is null")
			if err != nil {
				util.LogErr(err)
				return
			}
			defer rows.Close()

			messagesByLanguage := map[string][]string{}

			for rows.Next() {
				var firebaseToken string
				var language string
				if err := rows.Scan(&firebaseToken, &language); err != nil {
					util.LogErr(err)
					continue
				}
				_, err := dbConn.Exec(ctx, "UPDATE users SET special_offer_deadline = NOW() + INTERVAL '1 day', register_not = true WHERE firebase_token = $1", firebaseToken)
				if err != nil {
					util.LogErr(err)
				}
				messagesByLanguage[language] = append(messagesByLanguage[language], firebaseToken)
			}

			if err := rows.Err(); err != nil {
				util.LogErr(err)
			}

			for language, tokens := range messagesByLanguage {
				message := &messaging.MulticastMessage{
					Notification: &messaging.Notification{
						Title: util.GetTranslation(language, "special_offer_for_you_title"),
						Body:  util.GetTranslation(language, "special_offer_for_you_body"),
					},
					Data: map[string]string{
						"not_type": "show_paywall",
					},
					Tokens: tokens,
				}

				_, err = firebaseApp.SendEachForMulticast(context.Background(), message, 3)
				if err != nil {
					util.LogErr(err)
				} else {
					log.Printf("Sent non-premium push notification to %d users", len(tokens))
				}
			}
		}()
	}
}
