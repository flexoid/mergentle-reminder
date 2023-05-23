package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitLab struct {
		URL string `yaml:"url"`
	} `yaml:"gitlab"`
	Projects []ConfigProject
	Groups   []ConfigGroup
}

type ConfigGroup struct {
	ID int `yaml:"id"`
}

type ConfigProject struct {
	ID int `yaml:"id"`
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config, err := readConfig(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		os.Exit(1)
	}

	glClient, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"),
		gitlab.WithBaseURL(config.GitLab.URL))
	if err != nil {
		fmt.Printf("Error creating GitLab client: %v\n", err)
		os.Exit(1)
	}

	gitlabClient := &gitLabClient{client: glClient}

	mrs, err := fetchOpenedMergeRequests(config, gitlabClient)
	if err != nil {
		fmt.Printf("Error fetching opened merge requests: %v\n", err)
		os.Exit(1)
	}

	if len(mrs) == 0 {
		fmt.Println("No opened merge requests found.")
		os.Exit(0)
	}

	summary := formatMergeRequestsSummary(mrs)

	slackClient := &slackClient{webhookURL: os.Getenv("SLACK_WEBHOOK_URL")}
	err = sendSlackMessage(slackClient, summary)
	if err != nil {
		fmt.Printf("Error sending Slack message: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully sent merge request summary to Slack.")
}

func readConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
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

func formatMergeRequestsSummary(mrs []*MergeRequestWithApprovals) string {
	var summary string
	for _, mr := range mrs {
		approvedBy := strings.Join(mr.ApprovedBy, ", ")
		if approvedBy == "" {
			approvedBy = "None"
		}

		createdAtStr := mr.MergeRequest.CreatedAt.Format("2 January 2006, 15:04 MST")

		summary += fmt.Sprintf(
			":arrow_forward: <%s|%s>\n*Author:* %s\n*Created at:* %s\n*Approved by:* %s\n\n",
			mr.MergeRequest.WebURL, mr.MergeRequest.Title, mr.MergeRequest.Author.Name, createdAtStr, approvedBy,
		)
	}

	return summary
}
