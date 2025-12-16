package notification

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

type FCMService struct {
	Client *messaging.Client
}

func NewFCMService(credentialsFile string) *FCMService {
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("⚠️ Warning: Firebase Init Failed (Check credentials): %v", err)
		return &FCMService{Client: nil}
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("⚠️ Warning: FCM Client Init Failed: %v", err)
		return &FCMService{Client: nil}
	}

	log.Println("✅ Firebase FCM Connected")
	return &FCMService{Client: client}
}

// SendPush sends a message to multiple devices
func (s *FCMService) SendPush(tokens []string, title, body string, data map[string]string) {
	if s.Client == nil || len(tokens) == 0 {
		return
	}

	// Firebase allows sending up to 500 tokens at once
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data, // Custom data (e.g., group_id) to open the right chat screen
	}

	response, err := s.Client.SendMulticast(context.Background(), message)
	if err != nil {
		log.Printf("Error sending FCM: %v", err)
		return
	}

	if response.FailureCount > 0 {
		log.Printf("Sent %d messages, %d failed.", response.SuccessCount, response.FailureCount)
	}
}
