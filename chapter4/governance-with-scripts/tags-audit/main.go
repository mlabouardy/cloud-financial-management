package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type ResourceTagReport struct {
	TaggedResources    int
	UntaggedResources  int
	TagKeyValueCounter map[string]int
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	regions := getRegions(cfg)

	ec2Report := ResourceTagReport{TagKeyValueCounter: make(map[string]int)}

	for _, region := range regions {
		cfg.Region = region
		auditEC2(cfg, &ec2Report)
	}

	generatePieChart(ec2Report)
	printReport(ec2Report)
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

func auditEC2(cfg aws.Config, report *ResourceTagReport) {
	client := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{}
	output, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("Unable to describe EC2 instances, %v", err)
	}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if len(instance.Tags) == 0 {
				report.UntaggedResources++
			} else {
				report.TaggedResources++
				updateTagKeyValueCounter(instance.Tags, report)
			}
		}
	}
}

func updateTagKeyValueCounter(tags []types.Tag, report *ResourceTagReport) {
	for _, tag := range tags {
		keyValue := fmt.Sprintf("%s=%s", *tag.Key, *tag.Value)
		report.TagKeyValueCounter[keyValue]++
	}
}

func printReport(report ResourceTagReport) {
	fmt.Printf("Tagged Resources: %d\n", report.TaggedResources)
	fmt.Printf("Untagged Resources: %d\n\n", report.UntaggedResources)
	fmt.Println("Tag Key/Value pairs distribution:")
	for keyValue, count := range report.TagKeyValueCounter {
		fmt.Printf("%s: %d\n", keyValue, count)
	}
}

// Function to generate the pie chart using go-echarts
func generatePieChart(report ResourceTagReport) {
	// Create a new pie chart
	pie := charts.NewPie()

	// Prepare data for the pie chart
	pieItems := []opts.PieData{
		{Name: "Tagged", Value: report.TaggedResources},
		{Name: "Untagged", Value: report.UntaggedResources},
	}

	// Add data to the pie chart
	pie.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "Tags audit\n",
	}))

	pie.AddSeries("Resources", pieItems).
		SetSeriesOptions(charts.WithPieChartOpts(opts.PieChart{
			Radius: "50%",
		}))

	// Save the pie chart to an HTML file
	f, err := os.Create("ec2_tagged_vs_untagged_pie_chart.html")
	if err != nil {
		log.Fatalf("failed to create pie chart HTML file: %v", err)
	}
	defer f.Close()

	pie.Render(f)
}
