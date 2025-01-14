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

func handler() (map[string]interface{}, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	svc := ec2.NewFromConfig(cfg)

	resp, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalf("Failed to describe instances: %v", err)
	}

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceID := *instance.InstanceId
			tags := instance.Tags

			var correctedTags []types.Tag
			for _, tag := range tags {
				key := *tag.Key
				value := *tag.Value

				correctedKey := capitalize(key)
				correctedValue := capitalize(value)

				if key != correctedKey || value != correctedValue {
					correctedTags = append(correctedTags, types.Tag{
						Key:   aws.String(correctedKey),
						Value: aws.String(correctedValue),
					})
				}
			}

			if len(correctedTags) > 0 {
				_, err := svc.CreateTags(context.TODO(), &ec2.CreateTagsInput{
					Resources: []string{instanceID},
					Tags:      correctedTags,
				})
				if err != nil {
					log.Printf("Failed to update tags for instance %s: %v", instanceID, err)
				} else {
					fmt.Printf("Corrected tags for instance %s\n", instanceID)
				}
			}
		}
	}

	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Tag correction completed",
	}, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func main() {
	lambda.Start(handler)
}
