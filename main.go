package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/slack-go/slack"
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

type MergeRequestWithApprovals struct {
	MergeRequest *gitlab.MergeRequest
	ApprovedBy   []string
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
		options := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 50,
				Page:    1,
			},
		}

		for {
			projects, resp, err := client.Groups.ListGroupProjects(group.ID, options)
			if err != nil {
				return nil, err
			}

			for _, project := range projects {
				projectIDs = append(projectIDs, project.ID)
			}

			if resp.CurrentPage >= resp.TotalPages {
				break
			}

			options.Page = resp.NextPage
		}
	}

	return projectIDs, nil
}

func fetchOpenedMergeRequests(config *Config, client *gitlab.Client) ([]*MergeRequestWithApprovals, error) {
	var allMRs []*MergeRequestWithApprovals

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
			ListOptions: gitlab.ListOptions{
				PerPage: 50,
				Page:    1,
			},
		}

		for {
			mrs, resp, err := client.MergeRequests.ListProjectMergeRequests(projectID, options)
			if err != nil {
				return nil, err
			}

			for _, mr := range mrs {
				approvals, _, err := client.MergeRequestApprovals.GetConfiguration(projectID, mr.IID)
				if err != nil {
					return nil, err
				}

				approvedBy := make([]string, len(approvals.ApprovedBy))
				for i, approver := range approvals.ApprovedBy {
					approvedBy[i] = approver.User.Name
				}

				allMRs = append(allMRs, &MergeRequestWithApprovals{
					MergeRequest: mr,
					ApprovedBy:   approvedBy,
				})
			}

			if resp.CurrentPage >= resp.TotalPages {
				break
			}

			options.Page = resp.NextPage
		}
	}

	return allMRs, nil
}

func formatMergeRequestsSummary(mrs []*MergeRequestWithApprovals) string {
	var summary string
	for _, mr := range mrs {
		approvedBy := strings.Join(mr.ApprovedBy, ", ")
		if approvedBy == "" {
			approvedBy = "None"
		}

		summary += fmt.Sprintf(
			":arrow_forward: *Title:* <%s|%s>\n*Author:* %s\n*Created at:* %s\n*Approved by:* %s\n\n",
			mr.MergeRequest.WebURL, mr.MergeRequest.Title, mr.MergeRequest.Author.Name, mr.MergeRequest.CreatedAt, approvedBy,
		)
	}

	return summary
}

func sendSlackMessage(webhookURL, message string) error {
	msg := slack.WebhookMessage{
		Text: message,
	}
	return slack.PostWebhook(webhookURL, &msg)
}
