package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"sapps/lib/util"
	maindb "sapps/pkg/sapps/lib/db/main"
)

type RevenueCatService struct {
	db *maindb.MainDB
}

func NewRevenueCatService(db *maindb.MainDB) *RevenueCatService {
	return &RevenueCatService{
		db: db,
	}
}

type RevenueCatEvent struct {
	ID    *string `json:"id"`
	Event struct {
		ID                       *string  `json:"id"`
		AppUserID                string   `json:"app_user_id"`
		OriginalAppUserID        string   `json:"original_app_user_id"`
		ProductID                string   `json:"product_id"`
		NewProductID             string   `json:"new_product_id"`
		Price                    float64  `json:"price"`
		Currency                 string   `json:"currency"`
		PriceInPurchasedCurrency float64  `json:"price_in_purchased_currency"`
		TakehomePercentage       float64  `json:"takehome_percentage"`
		PurchasedAtMs            int64    `json:"purchased_at_ms"`
		ExpirationAtMs           int64    `json:"expiration_at_ms"`
		Store                    string   `json:"store"`
		Environment              string   `json:"environment"`
		TransactionID            string   `json:"transaction_id"`
		OriginalTransactionID    string   `json:"original_transaction_id"`
		Type                     string   `json:"type"`
		TransferredTo            []string `json:"transferred_to"`
		TransferredFrom          []string `json:"transferred_from"`
	} `json:"event"`
}

func (s *RevenueCatService) HandleWebhook(ctx context.Context, eventData *RevenueCatEvent) error {
	// Get or create event ID
	eventID := eventData.ID
	if eventData.ID == nil {
		eventID = eventData.Event.ID
	}
	if eventID == nil {
		tt := "tt"
		eventID = &tt
	}

	// Get product ID based on event type
	productID := eventData.Event.ProductID
	if eventData.Event.Type == "PRODUCT_CHANGE" {
		productID = eventData.Event.NewProductID
	}

	// Create log record
	otherData, err := json.Marshal(eventData.Event)
	if err != nil {
		util.LogErr(err)
		return err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO revenuecat_logs (
			revenuecat_event_id,
			app_user_id,
			original_app_user_id,
			product_id,
			price,
			currency,
			price_in_purchased_currency,
			takehome_percentage,
			purchased_at_ms,
			expiration_at_ms,
			store,
			environment,
			transaction_id,
			original_transaction_id,
			other_data,
			event_type,
			user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (revenuecat_event_id) DO NOTHING
	`,
		eventID,
		eventData.Event.AppUserID,
		eventData.Event.OriginalAppUserID,
		productID,
		eventData.Event.Price,
		eventData.Event.Currency,
		eventData.Event.PriceInPurchasedCurrency,
		eventData.Event.TakehomePercentage,
		eventData.Event.PurchasedAtMs,
		eventData.Event.ExpirationAtMs,
		eventData.Event.Store,
		eventData.Event.Environment,
		eventData.Event.TransactionID,
		eventData.Event.OriginalTransactionID,
		otherData,
		eventData.Event.Type,
		eventData.Event.AppUserID,
	)
	if err != nil {
		util.LogErr(err)
		return err
	}

	// Process based on event type
	switch eventData.Event.Type {
	case "INITIAL_PURCHASE", "RENEWAL", "CANCELLATION", "EXPIRATION", "PRODUCT_CHANGE", "NON_RENEWING_PURCHASE":
		return s.updatePremiumStatus(ctx, eventData.Event.AppUserID, eventData.Event.TransactionID, productID, eventData.Event.PurchasedAtMs, eventData.Event.ExpirationAtMs)
	case "TRANSFER":
		return s.handleTransfer(ctx, eventData)
	default:
		return nil
	}
}

func (s *RevenueCatService) handleTransfer(ctx context.Context, event *RevenueCatEvent) error {
	// End premium membership for old user
	transactionID := ""
	if len(event.Event.TransferredFrom) > 0 {
		for _, userID := range event.Event.TransferredFrom {
			var err error
			if transactionID, err = s.finishPremium(ctx, userID); err != nil {
				return err
			}
		}
	}
	if transactionID == "" {
		return errors.New("no transaction ID found")
	}
	// Start premium membership for new user
	if len(event.Event.TransferredTo) > 0 {
		for _, userID := range event.Event.TransferredTo {
			if err := s.updatePremiumStatus(ctx, userID, transactionID, "", event.Event.PurchasedAtMs, event.Event.ExpirationAtMs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RevenueCatService) updatePremiumStatus(ctx context.Context, userID, transactionID, productID string, startDateMs, endDateMs int64) error {
	startDate := time.UnixMilli(startDateMs)
	var endDate *time.Time
	if endDateMs > 0 {
		endDateMs := time.UnixMilli(endDateMs)
		endDate = &endDateMs
	}

	premiumType := productID
	if strings.Contains(premiumType, ":") {
		premiumType = strings.Split(premiumType, ":")[0]
	}

	// First, create or update premium_data
	_, err := s.db.Exec(ctx, `
		INSERT INTO premium_data (id, premium_type, created_date, expire_date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, transactionID, premiumType, startDate, endDate)
	if err != nil {
		util.LogErr(err)
		return err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE users 
		SET premium_id = NULL
		WHERE premium_id = $1
	`, transactionID)
	util.LogErr(err)

	// Then update user's premium_id
	_, err = s.db.Exec(ctx, `
		UPDATE users 
		SET premium_id = $1
		WHERE firebase_id = $2
	`, transactionID, userID)
	util.LogErr(err)
	return err
}

func (s *RevenueCatService) finishPremium(ctx context.Context, userID string) (string, error) {
	var transactionID string
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(premium_id, '') FROM users 
		WHERE firebase_id = $1
	`, userID).Scan(&transactionID)
	if err != nil {
		util.LogErr(err)
		return "", err
	}
	// Clear premium_id from user
	_, err = s.db.Exec(ctx, `
		UPDATE users 
		SET premium_id = NULL
		WHERE firebase_id = $1
	`, userID)
	if err != nil {
		util.LogErr(err)
		return "", err
	}
	return transactionID, nil
}
