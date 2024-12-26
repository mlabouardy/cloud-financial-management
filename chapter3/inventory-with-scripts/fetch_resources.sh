#!/bin/bash

# Define the output CSV file
output_file="aws_resources.csv"

# Add header to the CSV file
echo "Resource Name,Resource ARN,Service,Resource Region,Resource Creation Time,Resource Tags" > "$output_file"

# Get all AWS regions
regions=$(aws ec2 describe-regions --query "Regions[*].RegionName" --output text)

# Loop through each region and gather EC2, S3, and Lambda data
for region in $regions; do
  echo "Processing region: $region"

  # List EC2 instances in the region
  aws ec2 describe-instances --region "$region" --query "Reservations[*].Instances[*]" --output json | \
  jq -r --arg region "$region" '.[][] | 
  "\(.InstanceId),\(.InstanceId),EC2,\($region),\(.LaunchTime),\(.Tags | map(.Key + "=" + .Value) | join(";"))"' >> "$output_file"

  # List S3 buckets (S3 is global, so no region required for this call)
  if [ "$region" == "us-east-1" ]; then
    aws s3api list-buckets --query "Buckets[*]" --output json | \
    jq -r '.[] | 
    "\(.Name),arn:aws:s3:::\(.Name),S3,global,\(.CreationDate),"' >> "$output_file"
  fi

  # List Lambda functions in the region
  aws lambda list-functions --region "$region" --query "Functions[*]" --output json | \
  jq -r --arg region "$region" '.[] | 
  "\(.FunctionName),\(.FunctionArn),Lambda,\($region),\(.LastModified),"' >> "$output_file"
done

echo "AWS resource details saved to $output_file"
