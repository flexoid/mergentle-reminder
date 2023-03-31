package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitLab struct {
		URL string `yaml:"url"`
	} `yaml:"gitlab"`
	Projects []struct {
		ID int `yaml:"id"`
	} `yaml:"projects"`
	Groups []struct {
		ID int `yaml:"id"`
	} `yaml:"groups"`
}

func main() {
	config, err := readConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		os.Exit(1)
	}

	gitlabClient, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"),
		gitlab.WithBaseURL(config.GitLab.URL))
	if err != nil {
		fmt.Printf("Error creating GitLab client: %v\n", err)
		os.Exit(1)
	}

	mrs, err := fetchOpenedMergeRequests(config, gitlabClient)
	if err != nil {
		fmt.Printf("Error fetching opened merge requests: %v\n", err)
		os.Exit(1)
	}

	summary := formatMergeRequestsSummary(mrs)

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	err = sendSlackMessage(slackWebhookURL, summary)
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

func fetchProjectsFromGroups(config *Config, client *gitlab.Client) ([]int, error) {
	var projectIDs []int
	for _, group := range config.Groups {
		options := &gitlab.ListGroupProjectsOptions{}
		projects, _, err := client.Groups.ListGroupProjects(group.ID, options)
		if err != nil {
			return nil, err
		}

		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}
	}

	return projectIDs, nil
}

func fetchOpenedMergeRequests(config *Config, client *gitlab.Client) ([]*gitlab.MergeRequest, error) {
	var allMRs []*gitlab.MergeRequest

	// Add projects from groups to the projects list
	projectIDs, err := fetchProjectsFromGroups(config, client)
	if err != nil {
		return nil, err
	}

	for _, project := range config.Projects {
		projectIDs = append(projectIDs, project.ID)
	}

	for _, projectID := range projectIDs {
		options := &gitlab.ListProjectMergeRequestsOptions{
			State:   gitlab.String("opened"),
			OrderBy: gitlab.String("updated_at"),
			Sort:    gitlab.String("desc"),
			WIP:     gitlab.String("no"),
		}

		mrs, _, err := client.MergeRequests.ListProjectMergeRequests(projectID, options)
		if err != nil {
			return nil, err
		}

		allMRs = append(allMRs, mrs...)
	}

	return allMRs, nil
}

func formatMergeRequestsSummary(mrs []*gitlab.MergeRequest) string {
	var summary string
	for _, mr := range mrs {
		summary += fmt.Sprintf(
			"Title: %s\nURL: %s\nAuthor: %s\nCreated at: %s\nUpvotes: %d\nDownvotes: %d\nStatus: %s\n\n",
			mr.Title, mr.WebURL, mr.Author.Name, mr.CreatedAt, mr.Upvotes, mr.Downvotes, mr.State,
		)
	}

	return summary
}

func sendSlackMessage(webhookURL, message string) error {
	// msg := slack.WebhookMessage{
	// 	Text: message,
	// }
	// return slack.PostWebhook(webhookURL, &msg)
	fmt.Printf("Sending message to Slack:\n%s\n", message)
	return nil
}
