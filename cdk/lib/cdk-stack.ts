import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as events from "aws-cdk-lib/aws-events";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { GoFunction } from "@aws-cdk/aws-lambda-go-alpha";

export class CdkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // The Lambda function that will run the Go application.
    const mergentleReminderLambda = new GoFunction(
      this,
      "MergentleReminderLambda",
      {
        entry: "../", // The path to your Go module root
        runtime: lambda.Runtime.PROVIDED_AL2,
        architecture: lambda.Architecture.ARM_64,
        timeout: cdk.Duration.minutes(5),
        environment: {
          GITLAB_URL: process.env.GITLAB_URL || "",
          GITLAB_TOKEN: process.env.GITLAB_TOKEN || "",
          SLACK_WEBHOOK_URL: process.env.SLACK_WEBHOOK_URL || "",
          PROJECTS: process.env.PROJECTS || "",
          GROUPS: process.env.GROUPS || "",
          AUTHORS: process.env.AUTHORS || "",
        },
      }
    );

    // An EventBridge rule to run the Lambda function on a schedule.
    // Runs every weekday at 9am UTC. You can modify the schedule here.
    new events.Rule(this, "ScheduledRunRule", {
      schedule: events.Schedule.cron({
        minute: "0",
        hour: "7,13",
        weekDay: "MON-FRI",
      }),
      targets: [new targets.LambdaFunction(mergentleReminderLambda)],
    });
  }
}
