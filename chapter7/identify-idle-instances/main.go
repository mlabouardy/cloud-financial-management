package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type InstanceReport struct {
	InstanceID string
	Region     string
	CPUUtil    float64
	NetworkIn  float64
	NetworkOut float64
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	regions := getRegions(cfg)

	for _, region := range regions {
		cfg.Region = region
		instances := getUnderutilizedInstances(cfg, region)
		for _, instance := range instances {
			fmt.Printf("%s in %s: CPU %.2f%%, Network In %.2f KB, Network Out %.2f KB\n",
				instance.InstanceID, instance.Region, instance.CPUUtil, instance.NetworkIn, instance.NetworkOut)
		}
	}
}

func getRegions(cfg aws.Config) []string {
	ec2Client := ec2.NewFromConfig(cfg)
	output, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		log.Fatalf("Unable to describe regions, %v", err)
	}

	var regions []string
	for _, region := range output.Regions {
		regions = append(regions, *region.RegionName)
	}
	return regions
}

func getUnderutilizedInstances(cfg aws.Config, region string) []InstanceReport {
	ec2Client := ec2.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	input := &ec2.DescribeInstancesInput{}
	output, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("Unable to describe instances in %s, %v", region, err)
	}

	var results []InstanceReport

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			cpuUtil := getMetricAverage(cwClient, "CPUUtilization", *instance.InstanceId)
			networkIn := getMetricAverage(cwClient, "NetworkIn", *instance.InstanceId) / 1024
			networkOut := getMetricAverage(cwClient, "NetworkOut", *instance.InstanceId) / 1024

			if cpuUtil < 10 && networkIn < 1 && networkOut < 1 {
				results = append(results, InstanceReport{
					InstanceID: *instance.InstanceId,
					Region:     region,
					CPUUtil:    cpuUtil,
					NetworkIn:  networkIn,
					NetworkOut: networkOut,
				})
			}
		}
	}

	return results
}

func getMetricAverage(client *cloudwatch.Client, metricName, instanceID string) float64 {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/EC2"),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(instanceID),
			},
		},
		StartTime:  aws.Time(time.Now().Add(-7 * 24 * time.Hour)),
		EndTime:    aws.Time(time.Now()),
		Period:     aws.Int32(3600),
		Statistics: []types.Statistic{types.StatisticAverage},
	}

	output, err := client.GetMetricStatistics(context.TODO(), input)
	if err != nil || len(output.Datapoints) == 0 {
		return 0
	}

	return *output.Datapoints[0].Average
}
