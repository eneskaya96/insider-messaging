package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/persistence"
	"github.com/eneskaya/insider-messaging/pkg/config"
)

var (
	phoneNumbers = []string{
		"+905551111111", "+905552222222", "+905553333333", "+905554444444",
		"+905555555555", "+905556666666", "+905557777777", "+905558888888",
		"+905559999999", "+905550000000",
	}

	messageTemplates = []string{
		"Welcome to Insider! Your journey starts here.",
		"Special offer just for you! Check your account.",
		"Your verification code is: %d",
		"Thank you for choosing Insider!",
		"Limited time offer - Don't miss out!",
		"Your order has been confirmed.",
		"Meeting reminder: Team sync at %s",
		"System update completed successfully.",
		"New features available in your account.",
		"Your subscription has been renewed.",
		"Important: Security update required.",
		"Flash sale starts in 1 hour!",
		"Your feedback matters to us.",
		"Weekly summary: %d new updates",
		"Reminder: Complete your profile",
	}
)

func main() {
	log.Println("Starting database seeding...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := persistence.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := persistence.NewMessageRepositoryPostgres(db.DB(), cfg.Message.CharLimit)

	ctx := context.Background()
	messageCount := cfg.Seed.MessageCount

	log.Printf("Creating %d test messages...", messageCount)

	rand.Seed(time.Now().UnixNano())

	successCount := 0
	for i := 0; i < messageCount; i++ {
		phoneNumber := phoneNumbers[rand.Intn(len(phoneNumbers))]
		messageTemplate := messageTemplates[rand.Intn(len(messageTemplates))]

		content := fmt.Sprintf(messageTemplate, rand.Intn(10000))
		if len(content) > cfg.Message.CharLimit {
			content = content[:cfg.Message.CharLimit]
		}

		phone, err := valueobject.NewPhoneNumber(phoneNumber)
		if err != nil {
			log.Printf("Failed to create phone number: %v", err)
			continue
		}

		messageContent, err := valueobject.NewMessageContent(content, cfg.Message.CharLimit)
		if err != nil {
			log.Printf("Failed to create message content: %v", err)
			continue
		}

		message, err := entity.NewMessage(phone, messageContent, cfg.Message.MaxRetries)
		if err != nil {
			log.Printf("Failed to create message entity: %v", err)
			continue
		}

		if err := repo.Create(ctx, message); err != nil {
			log.Printf("Failed to save message: %v", err)
			continue
		}

		successCount++
		if (i+1)%10 == 0 {
			log.Printf("Progress: %d/%d messages created", successCount, messageCount)
		}
	}

	log.Printf("Seeding completed! Successfully created %d/%d messages", successCount, messageCount)
}
