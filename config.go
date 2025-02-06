package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GitLab struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"gitlab"`
	Slack struct {
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"slack"`
	Projects     []ConfigProject `yaml:"projects"`
	Groups       []ConfigGroup   `yaml:"groups"`
	CronSchedule string          `yaml:"cron_schedule"`
	Authors      []ConfigAuthor  `yaml:"authors"`
}

type ConfigGroup struct {
	ID int `yaml:"id"`
}

type ConfigProject struct {
	ID int `yaml:"id"`
}

type ConfigAuthor struct {
	ID       int    `yaml:"id"`
	Username string `yaml:"username"`
}

type Env interface {
	Getenv(key string) string
}

type OsEnv struct{}

func (e *OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

func loadConfig(env Env) (*Config, error) {
	configPath := env.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config := &Config{}
	var err error

	if _, err := os.Stat(configPath); err == nil {
		config, err = readConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("Error reading configuration file: %v\n", err)
		}
	}

	if env := env.Getenv("PROJECTS"); env != "" {
		config.Projects, err = parseIDsAsConfigProjects(env)
		if err != nil {
			return nil, fmt.Errorf("Error parsing PROJECTS environment variable: %v\n", err)
		}
	}

	if env := env.Getenv("GROUPS"); env != "" {
		config.Groups, err = parseIDsAsConfigGroups(env)
		if err != nil {
			return nil, fmt.Errorf("Error parsing GROUPS environment variable: %v\n", err)
		}
	}

	gitlabURL := env.Getenv("GITLAB_URL")
	if gitlabURL != "" {
		config.GitLab.URL = gitlabURL
	}
	if config.GitLab.URL == "" {
		config.GitLab.URL = "https://gitlab.com"
	}

	gitlabToken := env.Getenv("GITLAB_TOKEN")
	if gitlabToken != "" {
		config.GitLab.Token = gitlabToken
	}
	if config.GitLab.Token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN environment variable is required")
	}

	slackWebhookURL := env.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL != "" {
		config.Slack.WebhookURL = slackWebhookURL
	}
	if config.Slack.WebhookURL == "" {
		return nil, fmt.Errorf("SLACK_WEBHOOK_URL environment variable is required")
	}

	if env := env.Getenv("AUTHORS"); env != "" {
		config.Authors, err = parseAuthors(env)
		if err != nil {
			return nil, fmt.Errorf("Error parsing AUTHORS environment variable: %v\n", err)
		}
	}

	cronSchedule := env.Getenv("CRON_SCHEDULE")
	if cronSchedule != "" {
		config.CronSchedule = cronSchedule
	}

	if len(config.Projects) == 0 && len(config.Groups) == 0 {
		return nil, fmt.Errorf("Neither groups nor projects were provided")
	}

	return config, nil
}

func readConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func parseIDsAsConfigProjects(env string) ([]ConfigProject, error) {
	ids, err := parseIDs(env)
	if err != nil {
		return nil, err
	}

	var projects []ConfigProject
	for _, id := range ids {
		projects = append(projects, ConfigProject{ID: id})
	}

	return projects, nil
}

func parseIDsAsConfigGroups(env string) ([]ConfigGroup, error) {
	ids, err := parseIDs(env)
	if err != nil {
		return nil, err
	}

	var groups []ConfigGroup
	for _, id := range ids {
		groups = append(groups, ConfigGroup{ID: id})
	}

	return groups, nil
}

func parseIDs(env string) ([]int, error) {
	var ids []int
	for _, idStr := range strings.Split(env, ",") {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func parseAuthors(env string) ([]ConfigAuthor, error) {
	var authors []ConfigAuthor
	for _, authorStr := range strings.Split(env, ",") {
		author := ConfigAuthor{}
		if id, err := strconv.Atoi(authorStr); err == nil {
			author.ID = id
		} else {
			author.Username = authorStr
		}
		authors = append(authors, author)
	}
	return authors, nil
}
