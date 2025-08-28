package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

func handleRequest(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	rdsClient := rds.NewFromConfig(cfg)
	retentionDays := 7
	now := time.Now()

	// Delete old DB instance snapshots
	snapshots, err := rdsClient.DescribeDBSnapshots(ctx, &rds.DescribeDBSnapshotsInput{
		SnapshotType: aws.String("manual"),
	})
	if err != nil {
		log.Printf("Failed to list DB snapshots: %v", err)
	}

	for _, snap := range snapshots.DBSnapshots {
		if snap.SnapshotCreateTime != nil && now.Sub(*snap.SnapshotCreateTime).Hours() > float64(retentionDays*24) {
			_, err := rdsClient.DeleteDBSnapshot(ctx, &rds.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: snap.DBSnapshotIdentifier,
			})
			if err != nil {
				log.Printf("Error deleting snapshot %s: %v", *snap.DBSnapshotIdentifier, err)
			} else {
				log.Printf("Deleted snapshot: %s", *snap.DBSnapshotIdentifier)
			}
		}
	}

	// Delete old DB cluster snapshots
	clusterSnaps, err := rdsClient.DescribeDBClusterSnapshots(ctx, &rds.DescribeDBClusterSnapshotsInput{
		SnapshotType: aws.String("manual"),
	})
	if err != nil {
		log.Printf("Failed to list DB cluster snapshots: %v", err)
	}

	for _, snap := range clusterSnaps.DBClusterSnapshots {
		if snap.SnapshotCreateTime != nil && now.Sub(*snap.SnapshotCreateTime).Hours() > float64(retentionDays*24) {
			_, err := rdsClient.DeleteDBClusterSnapshot(ctx, &rds.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: snap.DBClusterSnapshotIdentifier,
			})
			if err != nil {
				log.Printf("Error deleting cluster snapshot %s: %v", *snap.DBClusterSnapshotIdentifier, err)
			} else {
				log.Printf("Deleted cluster snapshot: %s", *snap.DBClusterSnapshotIdentifier)
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
