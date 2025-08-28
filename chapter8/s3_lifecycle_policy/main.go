package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Error loading AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.PutBucketLifecycleConfiguration(context.TODO(), &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String("my-analytics-logs"),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: []types.LifecycleRule{
				{
					ID:     aws.String("TransitionLogsToIA"),
					Status: types.ExpirationStatusEnabled,
					Filter: &types.LifecycleRuleFilter{Prefix: aws.String("logs/")},
					Transitions: []types.Transition{
						{
							Days:         aws.Int32(30),
							StorageClass: types.TransitionStorageClassStandardIa,
						},
					},
					Expiration: &types.LifecycleExpiration{
						Days: aws.Int32(180),
					},
				},
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to apply lifecycle policy: %v", err)
	}

	log.Println("Lifecycle policy successfully applied to bucket")
}
