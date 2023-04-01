package main

import (
	"testing"

	"github.com/flexoid/mergentle-reminder/mocks"
	slack "github.com/slack-go/slack"
)

func TestSendSlackMessage(t *testing.T) {
	mockSlackClient := mocks.NewSlackClient(t)
	mockSlackClient.EXPECT().PostWebhook(&slack.WebhookMessage{Text: "hello"}).Return(nil)
	sendSlackMessage(mockSlackClient, "hello")
}
