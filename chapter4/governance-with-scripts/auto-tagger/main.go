package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2InstanceStateChangeEvent struct {
	Detail struct {
		InstanceID string `json:"instance-id"`
		State      string `json:"state"`
	} `json:"detail"`
}

// Function to handle the event
func handler(ctx context.Context, event EC2InstanceStateChangeEvent) error {
	if event.Detail.State == "running" {
		log.Printf("Instance %s is running, applying tags...", event.Detail.InstanceID)
		return applyTags(ctx, event.Detail.InstanceID)
	}

	log.Printf("Instance %s is in state %s, no tags applied.", event.Detail.InstanceID, event.Detail.State)
	return nil
}

// Function to apply tags to a given EC2 instance
func applyTags(ctx context.Context, instanceID string) error {
	// Load the default AWS SDK config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	// Create EC2 client
	svc := ec2.NewFromConfig(cfg)

	// Define the tags you want to apply
	tags := []types.Tag{
		{
			Key:   aws.String("Environment"),
			Value: aws.String("Production"),
		},
		{
			Key:   aws.String("Owner"),
			Value: aws.String("DevOps Team"),
		},
	}

	// Create input for CreateTags API call
	input := &ec2.CreateTagsInput{
		Resources: []string{instanceID},
		Tags:      tags,
	}

	// Apply tags to the EC2 instance
	_, err = svc.CreateTags(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to apply tags to instance %s: %v", instanceID, err)
	}

	log.Printf("Successfully applied tags to instance %s", instanceID)
	return nil
}

func main() {
	lambda.Start(handler)
}
