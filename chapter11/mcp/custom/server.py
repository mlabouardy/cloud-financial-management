import sys
import json
import boto3
from fastmcp import FastMCP

mcp = FastMCP("AWS Config NLQ with SQL")

config = boto3.client("config")
bedrock = boto3.client("bedrock-runtime")

SYSTEM_PROMPT = """
You are a FinOps SQL assistant for AWS Config.
You generate valid AWS Config SQL queries only.

Supported fields:
- resourceId
- resourceType
- awsRegion
- availabilityZone
- tags (via configuration.tags.key)
- configuration.complianceType (for rules)
- configuration.state
- relationships

Syntax rules:
- Use `SELECT field` or `SELECT COUNT(*)`
- Use `WHERE` for filters with =, IN, AND, OR
- Use `GROUP BY` to aggregate results (e.g., by region)
- Do NOT use table names
- Do NOT use JOINs

Only output the query. No explanation.
"""

def generate_sql_query(question):
    prompt = f"{SYSTEM_PROMPT}\n\nHuman: {question}\n\nAssistant:"
    
    response = bedrock.invoke_model(
        modelId="anthropic.claude-3-5-sonnet-20240620-v1:0",
        body=json.dumps({
            "anthropic_version": "bedrock-2023-05-31",
            "messages": [
                {"role": "user", "content": prompt}
            ],
            "max_tokens": 200,
            "temperature": 0.2,
        }),
        contentType="application/json",
        accept="application/json"
    )

    raw_body = response["body"].read().decode("utf-8")
    body = json.loads(raw_body)
    return body.get("completion", "").strip()


def run_aws_config_query(sql):
    query_id = config.start_config_query(Expression=sql)["QueryId"]
    # poll for completion
    while True:
        result = config.get_config_query_results(QueryId=query_id)
        if result["QueryStatus"] in ["COMPLETE", "FAILED"]:
            break
    if result["QueryStatus"] == "FAILED":
        raise Exception(f"Query failed: {result['QueryErrorMessage']}")
    return result["Results"]

@mcp.tool(name="aws-config-nlq")
def handle(question: str):
    sql = generate_sql_query(question)
    results = run_aws_config_query(sql)
    return {
        "query": sql,
        "results": results
    }

if __name__ == "__main__":
   mcp.run()