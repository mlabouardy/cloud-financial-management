package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type InstanceData struct {
	InstanceID   string
	Name         string
	Region       string
	InstanceType string
	Tags         map[string]string
	Cost         float64
}

func main() {
	err := fetchInstancesAndCosts()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func fetchInstancesAndCosts() error {
	ctx := context.TODO()

	// Load the default AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Get all regions
	ec2Client := ec2.NewFromConfig(cfg)
	regions, err := getAllRegions(ctx, ec2Client)
	if err != nil {
		return fmt.Errorf("unable to fetch regions, %v", err)
	}

	// Prepare CSV file
	file, err := os.Create("ec2_instances_costs.csv")
	if err != nil {
		return fmt.Errorf("unable to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"InstanceID", "Name", "Region", "InstanceType", "Tags", "Cost"})

	// Process instances for each region
	for _, region := range regions {
		cfg.Region = region
		ec2Client = ec2.NewFromConfig(cfg)
		ceClient := costexplorer.NewFromConfig(cfg)

		instances, err := describeInstancesInRegion(ctx, ec2Client, region)
		if err != nil {
			fmt.Printf("Error fetching instances for region %s: %v\n", region, err)
			continue
		}

		for _, instance := range instances {
			cost, err := getInstanceCost(ctx, ceClient, instance.InstanceID)
			if err != nil {
				fmt.Printf("Error fetching cost for instance %s in region %s: %v\n", instance.InstanceID, region, err)
				cost = 0.0
			}

			// Write to CSV
			writer.Write([]string{
				instance.InstanceID,
				instance.Name,
				instance.Region,
				instance.InstanceType,
				formatTags(instance.Tags),
				fmt.Sprintf("%.2f", cost),
			})
		}
	}

	fmt.Println("Instance data and costs have been written to ec2_instances_costs.csv")
	return nil
}

func getAllRegions(ctx context.Context, client *ec2.Client) ([]string, error) {
	// Use STS GetCallerIdentity to list regions
	input := &ec2.DescribeRegionsInput{}

	output, err := client.DescribeRegions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to describe regions: %v", err)
	}

	// Collect all region names
	var regions []string
	for _, region := range output.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}

	return regions, nil
}

func describeInstancesInRegion(ctx context.Context, ec2Client *ec2.Client, region string) ([]InstanceData, error) {
	input := &ec2.DescribeInstancesInput{}
	resp, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to describe instances, %v", err)
	}

	var instances []InstanceData
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			tags := make(map[string]string)
			var name string
			for _, tag := range instance.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
				}
			}

			instances = append(instances, InstanceData{
				InstanceID:   aws.ToString(instance.InstanceId),
				Name:         name,
				Region:       region,
				InstanceType: string(instance.InstanceType),
				Tags:         tags,
			})
		}
	}
	return instances, nil
}

func getInstanceCost(ctx context.Context, ceClient *costexplorer.Client, instanceID string) (float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	ceInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate.Format("2006-01-02")),
			End:   aws.String(endDate.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		Filter: &types.Expression{
			Tags: &types.TagValues{
				Key:    aws.String("RESOURCE_ID"),
				Values: []string{instanceID},
			},
		},
	}

	ceResp, err := ceClient.GetCostAndUsage(ctx, ceInput)
	if err != nil {
		return 0.0, fmt.Errorf("error fetching cost for instance %s: %v", instanceID, err)
	}

	// Parse cost
	var cost float64
	if len(ceResp.ResultsByTime) > 0 {
		amount := ceResp.ResultsByTime[0].Total["UnblendedCost"].Amount
		cost, err = strconv.ParseFloat(aws.ToString(amount), 64)
		if err != nil {
			return 0.0, fmt.Errorf("error parsing cost for instance %s: %v", instanceID, err)
		}
	}
	return cost, nil
}

func formatTags(tags map[string]string) string {
	var result string
	for k, v := range tags {
		result += fmt.Sprintf("%s=%s; ", k, v)
	}
	return result
}
