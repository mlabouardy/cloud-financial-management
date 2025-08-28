#!/bin/bash

SRC_PREFIX="s3://mlabouardy-billing-dummy/data-billing/billing-reports/"
DEST_PREFIX="s3://mlabouardy-billing-dummy/cur-flattened-cur/"

# Step 1: Find all .csv.gz file paths recursively
aws s3 ls "$SRC_PREFIX" --recursive | awk '{print $4}' | grep '\.csv\.gz$' | while read -r file_path; do
  echo "Copying $file_path..."
  aws s3 cp "s3://mlabouardy-billing-dummy/$file_path" "$DEST_PREFIX"
done

echo "✅ All .csv.gz files copied to $DEST_PREFIX"
