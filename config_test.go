package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockEnv struct {
	values map[string]string
}

func (e *MockEnv) Getenv(key string) string {
	return e.values[key]
}

func TestLoadConfig(t *testing.T) {
	// Test default values and required environment variables
	t.Run("missing required env variables", func(t *testing.T) {
		env := &MockEnv{values: map[string]string{}}
		_, err := loadConfig(env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GITLAB_TOKEN environment variable is required")
	})

	t.Run("default GitLab URL", func(t *testing.T) {
		env := &MockEnv{values: map[string]string{
			"GITLAB_TOKEN":      "token",
			"SLACK_WEBHOOK_URL": "webhook",
			"CONFIG_PATH":       "NONEXISTING.yaml",
			"PROJECTS":          "1,2,3",
		}}

		config, err := loadConfig(env)
		assert.NoError(t, err)
		assert.Equal(t, "https://gitlab.com", config.GitLab.URL)
	})

	// Test overriding default values with environment variables
	t.Run("env variables overriding defaults", func(t *testing.T) {
		env := &MockEnv{values: map[string]string{
			"GITLAB_URL":        "https://gitlab.example.com",
			"GITLAB_TOKEN":      "token",
			"SLACK_WEBHOOK_URL": "webhook",
			"CONFIG_PATH":       "NONEXISTING.yaml",
			"PROJECTS":          "1,2,3",
		}}

		config, err := loadConfig(env)
		assert.NoError(t, err)
		assert.Equal(t, "https://gitlab.example.com", config.GitLab.URL)
		assert.Equal(t, []ConfigProject{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		}, config.Projects)
	})

	// Test loading config from file
	t.Run("loading from config file", func(t *testing.T) {
		env := &MockEnv{values: map[string]string{
			"CONFIG_PATH": "config.test.yaml",
		}}

		config, err := loadConfig(env)
		assert.NoError(t, err)

		assert.Equal(t, "https://gitlab.example.com", config.GitLab.URL)
		assert.Equal(t, "abcdef1234567890", config.GitLab.Token)
		assert.Equal(t, "https://hooks.slack.com/services/your-slack-webhook-url", config.Slack.WebhookURL)
		assert.Equal(t, []ConfigProject{
			{ID: 123},
			{ID: 456},
		}, config.Projects)
		assert.Equal(t, []ConfigGroup{
			{ID: 1},
			{ID: 2},
		}, config.Groups)
	})
}
