package export

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/cloudasset/v1"
)

// The dataset ID in BigQuery where the assets will be exported.
const (
	projectID = "your-project-id" // Replace with your GCP project ID
	datasetID = "your-dataset-id" // Replace with your BigQuery dataset ID
)

// CloudAssetExport is triggered by a Pub/Sub message.
func CloudAssetExport(ctx context.Context, _ interface{}) error {
	// 1. Create a timestamped table name with format: exported_data_YYYYMMDD_HHMMSS
	tableName := fmt.Sprintf("exported_data_%s", time.Now().Format("20060102_150405"))

	// 2. Set up the Cloud Asset API client
	assetService, err := cloudasset.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create cloudasset service: %v", err)
		return err
	}

	// 3. Configure the export request to BigQuery
	exportRequest := &cloudasset.ExportAssetsRequest{
		// Set the output config to use BigQuery
		OutputConfig: &cloudasset.OutputConfig{
			BigqueryDestination: &cloudasset.BigQueryDestination{
				Dataset: fmt.Sprintf("projects/%s/datasets/%s", projectID, datasetID),
				Table:   tableName,
				Force:   true, // Overwrite table if it exists
				PartitionSpec: &cloudasset.PartitionSpec{
					PartitionKey: "READ_TIME", // You can set partition key for BigQuery
				},
			},
		},
	}

	// 4. Set up the export operation
	exportAssetsCall := assetService.V1.ExportAssets(fmt.Sprintf("projects/%s", projectID), exportRequest)

	// 5. Execute the export
	_, err = exportAssetsCall.Do()
	if err != nil {
		log.Fatalf("Failed to export assets: %v", err)
		return err
	}

	log.Printf("Cloud assets successfully exported to BigQuery table: %s.%s", datasetID, tableName)
	return nil
}
