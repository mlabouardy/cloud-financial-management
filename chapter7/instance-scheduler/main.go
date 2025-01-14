package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func stopInstances(client *ec2.Client, instanceIDs []string) {
	if len(instanceIDs) == 0 {
		log.Println("No instances to stop")
		return
	}
	_, err := client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		log.Fatalf("Failed to stop instances: %v", err)
	}
	fmt.Printf("Stopped instances: %v\n", instanceIDs)
}

func startInstances(client *ec2.Client, instanceIDs []string) {
	if len(instanceIDs) == 0 {
		log.Println("No instances to start")
		return
	}
	_, err := client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		log.Fatalf("Failed to start instances: %v", err)
	}
	fmt.Printf("Started instances: %v\n", instanceIDs)
}

func handler(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	action := os.Getenv("ACTION") // 'stop' or 'start'
	ec2Client := ec2.NewFromConfig(cfg)

	resp, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Environment"),
				Values: []string{"Development"},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to describe instances: %v", err)
	}

	var instanceIDs []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	if action == "stop" {
		stopInstances(ec2Client, instanceIDs)
	} else if action == "start" {
		startInstances(ec2Client, instanceIDs)
	} else {
		log.Fatalf("Invalid action: %s", action)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
