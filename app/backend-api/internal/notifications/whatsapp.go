package notifications

import (
	"context"
	"fmt"
)

type WhatsAppSender struct{}

func NewWhatsAppSender() *WhatsAppSender {
	return &WhatsAppSender{}
}

func (s *WhatsAppSender) Send(ctx context.Context, phone, msg string) error {
	fmt.Println("SENDING WHATSAPP TO...", phone, msg)
	return nil
}
