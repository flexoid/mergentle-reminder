package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/reugn/go-quartz/job"
	"github.com/reugn/go-quartz/quartz"
	"github.com/xanzy/go-gitlab"
)

func main() {
	// Load configuration
	config, err := loadConfig(&OsEnv{})
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		os.Exit(1)
	}

	if config.CronSchedule == "" {
		log.Printf("Running in one-shot mode")
		execute(config)
		return
	}

	runScheduler(config)
}

func runScheduler(config *Config) {
	log.Printf("Running in cron mode with schedule: %s\n", config.CronSchedule)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := quartz.NewStdScheduler()
	sched.Start(ctx)

	cronTrigger, err := quartz.NewCronTrigger(config.CronSchedule)
	if err != nil {
		log.Printf("Error creating cron trigger: %v\n", err)
		os.Exit(1)
	}

	executeJob := job.NewFunctionJob(func(_ context.Context) (int, error) {
		execute(config)
		return 0, nil
	})

	err = sched.ScheduleJob(quartz.NewJobDetail(executeJob, quartz.NewJobKey("executeJob")), cronTrigger)
	if err != nil {
		log.Printf("Error scheduling job: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
}

func execute(config *Config) {
	glClient, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"),
		gitlab.WithBaseURL(config.GitLab.URL))
	if err != nil {
		log.Printf("Error creating GitLab client: %v\n", err)
		os.Exit(1)
	}

	gitlabClient := &gitLabClient{client: glClient}

	mrs, err := fetchOpenedMergeRequests(config, gitlabClient)
	if err != nil {
		log.Printf("Error fetching opened merge requests: %v\n", err)
		os.Exit(1)
	}

	if len(mrs) == 0 {
		log.Println("No opened merge requests found.")
		os.Exit(0)
	}

	summary := formatMergeRequestsSummary(mrs)

	slackClient := &slackClient{webhookURL: os.Getenv("SLACK_WEBHOOK_URL")}
	err = sendSlackMessage(slackClient, summary)
	if err != nil {
		log.Printf("Error sending Slack message: %v\n", err)
		os.Exit(1)
	}

	log.Println("Successfully sent merge request summary to Slack.")
}

func formatMergeRequestsSummary(mrs []*MergeRequestWithApprovals) string {
	var summary string
	for _, mr := range mrs {
		approvedBy := strings.Join(mr.ApprovedBy, ", ")
		if approvedBy == "" {
			approvedBy = "None"
		}

		createdAtStr := mr.MergeRequest.CreatedAt.Format("2 January 2006, 15:04 MST")

		var extra string
		if !mr.MergeRequest.BlockingDiscussionsResolved {
			extra = ":warning: Has unresolved blocking discussions"
		}

		summary += fmt.Sprintf(
			":arrow_forward: <%s|%s>\n*Author:* %s\n*Created at:* %s\n*Approved by:* %s\n",
			mr.MergeRequest.WebURL, mr.MergeRequest.Title, mr.MergeRequest.Author.Name, createdAtStr, approvedBy,
		)

		if extra != "" {
			summary += fmt.Sprintf("*Extra:* %s\n", extra)
		}

		summary += "\n"
	}

	return summary
}
