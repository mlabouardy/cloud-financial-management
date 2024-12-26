#!/bin/bash

# Define the output CSV file
output_file="ec2_instances.csv"

# Add header to the CSV file
echo "Instance ID,Instance Type,Region,Launch Time,Instance State,Tags" > "$output_file"

# Get all AWS regions
regions=$(aws ec2 describe-regions --query "Regions[*].RegionName" --output text)

# Loop through each region
for region in $regions; do
  echo "Processing region: $region"

  # Describe EC2 instances in the region
  aws ec2 describe-instances --region "$region" --query "Reservations[*].Instances[*]" --output json | \
  jq -r --arg region "$region" '.[][] | 
  "\(.InstanceId),\(.InstanceType),\($region),\(.LaunchTime),\(.State.Name),\(.Tags | map(.Key + "=" + .Value) | join(";"))"' >> "$output_file"
done

echo "EC2 instance details saved to $output_file"