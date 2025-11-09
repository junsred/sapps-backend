package connection

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const (
	minBackoff = 100 * time.Millisecond
	maxBackoff = 1 * time.Minute
	projectID  = "face-c3347"
)

type firebaseHolder struct {
	init sync.Once
	a    *FirebaseApp
}

var singletonFirebase firebaseHolder

func (h *firebaseHolder) setup() *FirebaseApp {
	h.init.Do(func() {
		s := setupFirebase()
		h.a = s
	})
	return h.a
}

func InjectFirebase() *FirebaseApp {
	return singletonFirebase.setup()
}

func setupFirebase() *FirebaseApp {
	fcmClient, err := NewFCMClient(os.Getenv("FCM_CREDENTIALS_PATH"))
	if err != nil {
		log.Fatalln(err)
	}
	return fcmClient
}

func retry(fn func() error, attempts int) error {
	var attempt int
	for {
		err := fn()
		if err == nil {
			return nil
		}

		if tErr, ok := err.(net.Error); !ok || !tErr.Timeout() {
			return err
		}

		attempt++
		backoff := minBackoff * time.Duration(attempt*attempt)
		if attempt > attempts || backoff > maxBackoff {
			return err
		}

		time.Sleep(backoff)
	}
}

var (
	ErrInvalidAPIKey = errors.New("client API Key is invalid")
)

type FirebaseApp struct {
	app       *firebase.App
	fcmClient *messaging.Client
	auth      *auth.Client
}

func NewFCMClient(credentialsFile string) (*FirebaseApp, error) {
	if credentialsFile == "" {
		return nil, ErrInvalidAPIKey
	}
	ctx := context.Background()
	cc, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	}, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	fcmClient, err := cc.Messaging(ctx)
	if err != nil {
		return nil, err
	}
	authClient, err := cc.Auth(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("FIREBASE CONNECTED")
	return &FirebaseApp{
		app:       cc,
		fcmClient: fcmClient,
		auth:      authClient,
	}, nil
}

func (c *FirebaseApp) SendWithRetry(ctx context.Context, msg *messaging.Message, retryAttempts int) (*messaging.BatchResponse, error) {
	resp := new(messaging.BatchResponse)
	err := retry(func() error {
		var er error
		resp, er = c.fcmClient.SendEach(ctx, []*messaging.Message{msg})
		if er != nil {
			return er
		}
		for _, r := range resp.Responses {
			if r.Error != nil {
				return r.Error
			}
		}
		return nil
	}, retryAttempts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *FirebaseApp) SendEachForMulticast(ctx context.Context, multicastMessage *messaging.MulticastMessage, retryAttempts int) (*messaging.BatchResponse, error) {
	resp := new(messaging.BatchResponse)
	err := retry(func() error {
		var er error
		resp, er = c.fcmClient.SendEachForMulticast(ctx, multicastMessage)
		if er != nil {
			return er
		}
		for _, r := range resp.Responses {
			if r.Error != nil {
				return r.Error
			}
		}
		return nil
	}, retryAttempts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *FirebaseApp) AuthToken(ctx context.Context, token string) (*auth.Token, error) {
	tk, err := c.auth.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return tk, nil
}

func (c *FirebaseApp) DeleteUser(ctx context.Context, uid string) error {
	err := c.auth.DeleteUser(ctx, uid)
	return err
}

func (c *FirebaseApp) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := c.auth.GetUser(ctx, uid)
	return user, err
}
