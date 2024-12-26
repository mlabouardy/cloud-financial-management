
  create view "asset_inventory"."public"."aws_resources__dbt_tmp"
    
    
  as (
    
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

region
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_apigateway_rest_api_stages' AS _cq_table
FROM aws_apigateway_rest_api_stages
UNION ALL 
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

region
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_apigateway_rest_apis' AS _cq_table
FROM aws_apigateway_rest_apis
UNION ALL 
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

'unavailable'
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_cloudfront_distributions' AS _cq_table
FROM aws_cloudfront_distributions
UNION ALL 
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

region
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_ec2_instances' AS _cq_table
FROM aws_ec2_instances
UNION ALL 
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

region
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_ec2_security_groups' AS _cq_table
FROM aws_ec2_security_groups
UNION ALL 
SELECT
_cq_id, _cq_source_name, _cq_sync_time,

account_id
 AS account_id,

SPLIT_PART(arn, ':', 5)
 AS request_account_id, 
CASE
WHEN SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 2) = '' AND SPLIT_PART(arn, ':', 7) = '' THEN NULL
ELSE SPLIT_PART(SPLIT_PART(arn, ':', 6), '/', 1)
END AS TYPE,
arn,

region
 AS region,

tags
 AS tags,
SPLIT_PART(arn, ':', 2) AS PARTITION,
SPLIT_PART(arn, ':', 3) AS service,
'aws_s3_buckets' AS _cq_table
FROM aws_s3_buckets

  );