import boto3
import gzip
import json
import pandas as pd
from io import BytesIO

# Configuration
region = "eu-central-1"
bucket = "mlabouardy-billing-dummy"
key = "data/billing/20250701-20250801/20250701T124318Z/billing-00001.csv.gz"
model_id = "anthropic.claude-3-5-sonnet-20240620-v1:0"

# Step 1: Download and decompress CSV from S3
s3 = boto3.client("s3", region_name=region)
obj = s3.get_object(Bucket=bucket, Key=key)
with gzip.GzipFile(fileobj=obj["Body"]) as gz:
    df = pd.read_csv(gz)

# --- Preprocess the CSV ---
grouped = df.groupby("lineItem/ProductCode")["lineItem/UnblendedCost"].sum()
grouped = grouped.sort_values(ascending=False).head(10)
summary = "\n".join([f"{k}: ${v:,.2f}" for k, v in grouped.items()])

# Optional: Add total
total = df["lineItem/UnblendedCost"].sum()

prompt = f"""
Here is a summary of the top AWS services from a Cost and Usage Report:

{summary}

Total monthly spend: ${total:,.2f}

Please write a financial summary: highlight top spending categories, detect anomalies if any, and suggest optimizations.
"""

payload = {
    "anthropic_version": "bedrock-2023-05-31",
    "messages": [
        {"role": "user", "content": prompt}
    ],
    "max_tokens": 200000,
    "temperature": 0.5
}

# Step 4: Call Claude 3.5 Sonnet via Bedrock
bedrock = boto3.client("bedrock-runtime", region_name=region)

response = bedrock.invoke_model(
    modelId=model_id,
    body=json.dumps(payload),
    contentType="application/json",
    accept="application/json"
)

# Step 5: Print Claude's summary
response_body = json.loads(response["body"].read())
print(response_body["content"][0]["text"])

