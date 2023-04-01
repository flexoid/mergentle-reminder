package main

import "github.com/slack-go/slack"

//go:generate mockery --name SlackClient
type SlackClient interface {
	PostWebhook(payload *slack.WebhookMessage) error
}

type slackClient struct {
	webhookURL string
}

func (c *slackClient) PostWebhook(payload *slack.WebhookMessage) error {
	return slack.PostWebhook(c.webhookURL, payload)
}

func sendSlackMessage(client SlackClient, message string) error {
	msg := slack.WebhookMessage{
		Text: message,
	}
	return client.PostWebhook(&msg)
}
