package main

import (
	"context"
	"sapps/lib/connection"
	"sapps/lib/util"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"firebase.google.com/go/v4/messaging"
)

func main() {
	// Check command line arguments for soulmate notification
	SendPushToNonPremiumUsers()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Server Exited Properly")
}

func SendPushToNonPremiumUsers() {
	firebaseApp := connection.InjectFirebase()
	dbConn := connection.InjectMainDB()
	ctx := context.Background()

	rows, err := dbConn.Query(ctx, "SELECT firebase_token, language FROM users WHERE premium_id IS NULL AND firebase_token IS NOT NULL AND firebase_token != '' AND language IS NOT NULL")
	if err != nil {
		util.LogErr(err)
		return
	}
	defer rows.Close()

	tokensByLanguage := make(map[string][]string)
	for rows.Next() {
		var firebaseToken *string
		var language *string
		if err := rows.Scan(&firebaseToken, &language); err != nil {
			util.LogErr(err)
			continue
		}
		if firebaseToken == nil || language == nil {
			continue
		}
		tokensByLanguage[*language] = append(tokensByLanguage[*language], *firebaseToken)
		_, err := dbConn.Exec(ctx, "UPDATE users SET special_offer_deadline = NOW() + INTERVAL '1 day' WHERE firebase_token = $1", firebaseToken)
		if err != nil {
			util.LogErr(err)
		}
	}

	if err := rows.Err(); err != nil {
		util.LogErr(err)
	}
	//print(util.GetTranslation("en", "special_offer_for_you_title_97_off"))
	//print(util.GetTranslation("en", "special_offer_for_you_body_97_off"))
	for language, tokens := range tokensByLanguage {
		// 500 buckets
		for i := 0; i < len(tokens); i += 500 {
			batch := tokens[i:min(i+500, len(tokens))]
			message := &messaging.MulticastMessage{
				Notification: &messaging.Notification{
					Title: "PRO 87% OFF",
					Body:  "Limited time offer!",
				},
				Data: map[string]string{
					"not_type": "show_paywall",
				},
				Tokens: batch,
			}

			_, err = firebaseApp.SendEachForMulticast(context.Background(), message, 3)
			if err != nil {
				util.LogErr(err)
			}
			log.Printf("Sent non-premium push notification to %d %s users", len(batch), language)
		}
	}
}

func init() {
	util.LoadFolder()
}
