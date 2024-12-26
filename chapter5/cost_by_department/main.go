package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func main() {
	// Load the AWS configuration (from environment, shared config, etc.)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create a Cost Explorer client
	client := costexplorer.NewFromConfig(cfg)

	// Define the start and end dates for the last 30 days
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	// Format the dates as required by AWS Cost Explorer (YYYY-MM-DD)
	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")

	// Define the time period and grouping options
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeTag,
				Key:  aws.String("Project"),
			},
		},
	}

	// Call the Cost Explorer API to get cost data
	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to get cost and usage: %v", err)
	}

	// Process and display the results
	fmt.Printf("Cost data for the last 30 days (grouped by Project):\n")
	for _, group := range result.ResultsByTime {
		fmt.Printf("Time period: %s to %s\n", *group.TimePeriod.Start, *group.TimePeriod.End)
		for _, groupRow := range group.Groups {
			projectTag := (groupRow.Keys)[0]
			cost := groupRow.Metrics["UnblendedCost"].Amount
			fmt.Printf("Project: %s, Cost: $%s\n", projectTag, *cost)
		}
	}
}
