#!/bin/bash
set -euo pipefail

output_file="aws_resources.csv"  #A
echo "Resource Name,Resource ARN,Service,Resource Region,Resource Creation Time,Resource Tags" > "$output_file"  #B

# Helper: join S3 TagSet JSON -> "k=v;k=v" (empty if none)
s3_tags() {
  local bucket="$1"
  # get-bucket-tagging fails if no tags — swallow errors and output empty
  aws s3api get-bucket-tagging --bucket "$bucket" --output json 2>/dev/null \
    | jq -r '(.TagSet // []) | map("\(.Key)=\(.Value)") | join(";")'
}

# Helper: join Lambda tag map -> "k=v;k=v" (empty if none)
lambda_tags() {
  local arn="$1"
  aws lambda list-tags --resource "$arn" --output json 2>/dev/null \
    | jq -r '(.Tags // {}) | to_entries | map("\(.key)=\(.value)") | join(";")'
}

regions=$(aws ec2 describe-regions --query "Regions[*].RegionName" --output text)  #C

for region in $regions; do  #D
  echo "Processing region: $region"

  # ------- EC2 (tags inline) -------
  aws ec2 describe-instances \
    --region "$region" \
    --query "Reservations[*].Instances[*]" \
    --output json \
  | jq -r --arg region "$region" '.[][] |
      "\(.InstanceId),\(.InstanceId),EC2,\($region),\(.LaunchTime),\((.Tags // []) | map(.Key + "=" + .Value) | join(";"))"' \
  >> "$output_file"  #E

  # ------- S3 (global listing, tag per bucket) -------
  if [ "$region" == "us-east-1" ]; then
    # Name + CreationDate from the global call
    while IFS=$'\t' read -r name creation; do
      tags="$(s3_tags "$name")"
      printf '%s\n' \
        "${name},arn:aws:s3:::${name},S3,global,${creation},${tags}" \
      >> "$output_file"
    done < <(aws s3api list-buckets --query "Buckets[*].[Name,CreationDate]" --output text)
  fi

  # ------- Lambda (per region, fetch tags per function) -------
  # Pull a compact list to minimize parsing overhead
  while IFS=$'\t' read -r fname arn lastmod; do
    ltags="$(lambda_tags "$arn")"
    printf '%s\n' \
      "${fname},${arn},Lambda,${region},${lastmod},${ltags}" \
    >> "$output_file"
  done < <(aws lambda list-functions \
            --region "$region" \
            --query "Functions[*].[FunctionName,FunctionArn,LastModified]" \
            --output text)
done

echo "AWS resource details saved to $output_file"