### Calculating Costs by Instance Type

This query calculates the total cost of active EC2 instances grouped by instance type. It helps you identify which instance types incur the highest costs, allowing you to evaluate whether rightsizing or switching to a different pricing model could reduce expenses.


#### Listing 7.1 Query to calculate cost by EC2 instance type


```
SELECT 
    product_instance_type AS instance_type, 
    SUM(line_item_unblended_cost) AS total_cost 
FROM 
    cur_db.data 
WHERE 
    line_item_product_code = 'AmazonEC2'
    AND
    line_item_line_item_type = 'Usage'
GROUP BY 
    product_instance_type;
```


The query filters the CUR dataset for EC2 usage (*line_item_product_code = 'AmazonEC2'*), groups the data by instance type (*product_instance_type*), and calculates the unblended cost (raw cost without discounts).


### Analyzing Costs by Purchase Option

This query breaks down costs by EC2 purchase options (On-Demand, Spot, Savings Plan, Reserved Instances) and shows how much each type contributes to overall spend. This breakdown helps identify where cost savings opportunities lie—for example, underutilized Reserved Instances or excess Spot usage.


#### Listing 7.2 Query to calculate cost by EC2 purchase options


```
SELECT 
  bill_billing_period_start_date,
  line_item_usage_start_date, 
  bill_payer_account_id, 
  line_item_usage_account_id,
  CASE 
    WHEN (line_item_usage_type LIKE '%SpotUsage%') THEN SPLIT_PART(line_item_usage_type, ':', 2)
    ELSE product_instance_type
  END AS case_product_instance_type,
  CASE
    WHEN (savings_plan_savings_plan_a_r_n <> '') THEN 'SavingsPlan'
    WHEN (reservation_reservation_a_r_n <> '') THEN 'Reserved'
    WHEN (line_item_usage_type LIKE '%Spot%') THEN 'Spot'
    ELSE 'OnDemand' 
  END AS case_purchase_option, 
  SUM(CASE
    WHEN line_item_line_item_type = 'SavingsPlanCoveredUsage' THEN savings_plan_savings_plan_effective_cost
    WHEN line_item_line_item_type = 'DiscountedUsage' THEN reservation_effective_cost
    WHEN line_item_line_item_type = 'Usage' THEN line_item_unblended_cost
    ELSE 0 
  END) AS sum_amortized_cost, 
  SUM(line_item_usage_amount) AS sum_line_item_usage_amount
FROM 
  cur_db.data  
WHERE 
  line_item_usage_start_date >= date_add('month', -6, current_date)
  AND (line_item_product_code = 'AmazonEC2'
    AND product_servicecode <> 'AWSDataTransfer'
    AND line_item_operation LIKE '%RunInstances%'
    AND line_item_usage_type NOT LIKE '%DataXfer%'
  )
  AND (line_item_line_item_type = 'Usage'
    OR (line_item_line_item_type = 'SavingsPlanCoveredUsage')
    OR (line_item_line_item_type = 'DiscountedUsage')
  )
  -- excludes consumed ODCR hours from total
  AND product['capacitystatus'] != 'AllocatedCapacityReservation'
GROUP BY 
  bill_billing_period_start_date,
  line_item_usage_start_date, 
  bill_payer_account_id, 
  line_item_usage_account_id,
  5, --refers to case_product_instance_type
  6 --refers to case_purchase_option 
ORDER BY 
  sum_line_item_usage_amount DESC;
```



### Identifying Underutilized Reserved Instances

This query pulls all active Reserved Instance ARNs for Amazon EC2 and produces their utilization for last month. This will give you a very granular look at which Reserved Instance purchases were not being utilized to their full extent last month.


#### Listing 7.3 Query calculating the RIs utilization for last month


```
SELECT
  bill_payer_account_id,
  line_item_usage_account_id,
  DATE_FORMAT((line_item_usage_start_date),'%Y-%m') AS month_line_item_usage_start_date,
  bill_bill_type,
  line_item_product_code,
  line_item_usage_type,
  product['region'],
  reservation_subscription_id,
  reservation_reservation_a_r_n,
  pricing_purchase_option,
  pricing_offering_class,
  pricing_lease_contract_length,
  reservation_number_of_reservations,
  reservation_start_time,
  reservation_end_time,
  reservation_modification_status,
  reservation_total_reserved_units,
  reservation_unused_quantity,
  TRY_CAST(1 - (TRY_CAST(reservation_unused_quantity AS DECIMAL(16,8)) / TRY_CAST(reservation_total_reserved_units AS DECIMAL(16,8))) as DECIMAL(16,8)) AS calc_percentage_utilized
FROM
  cur_db.data
WHERE 
  DATE_TRUNC('month', line_item_usage_start_date) = "date_trunc"('month', current_date) - INTERVAL  '1' MONTH --last month
  AND pricing_term = 'Reserved'
  AND line_item_line_item_type IN ('Fee','RIFee')
  AND line_item_product_code = 'AmazonEC2' --EC2 only, comment out for all reservation types
  AND bill_bill_type = 'Anniversary' --identify 
  AND try_cast(date_parse(SPLIT_PART(reservation_end_time, 'T', 1), '%Y-%m-%d') as date) > cast(current_date as date) --res exp time after today's date
GROUP BY 
  bill_bill_type,
  bill_payer_account_id,
  line_item_usage_account_id,
  reservation_reservation_a_r_n,
  reservation_subscription_id,
  DATE_FORMAT((line_item_usage_start_date),'%Y-%m'),
  line_item_product_code,
  line_item_usage_type,
  product['region'],
  pricing_purchase_option,
  pricing_offering_class,
  pricing_lease_contract_length,
  reservation_number_of_reservations,
  reservation_start_time,
  reservation_end_time,
  reservation_modification_status,
  reservation_total_reserved_units,
  reservation_unused_quantity
ORDER BY 
  reservation_unused_quantity DESC,
  reservation_end_time ASC,
  calc_percentage_utilized ASC;
```



### Analyzing AWS Lambda Costs

AWS Lambda incurs costs based on factors like requests, execution time, and data transfer. This query breaks down Lambda costs by usage type, resource ID (Function ARN), and pricing plan, providing a granular view of where expenses originate.


#### Listing 7.4 Query listing all Lambda functions with their associated cost and usage


```
SELECT *
FROM
(
  (  
    SELECT
      bill_payer_account_id,
      line_item_usage_account_id, 
      line_item_line_item_type,
      DATE_FORMAT((line_item_usage_start_date),'%Y-%m-%d') AS day_line_item_usage_start_date,
      product['region'],
      CASE
        WHEN line_item_usage_type LIKE '%Lambda-Edge-GB-Second%' THEN 'Lambda EDGE GB x Sec.'
        WHEN line_item_usage_type LIKE '%Lambda-Edge-Request%' THEN 'Lambda EDGE Requests'
        WHEN line_item_usage_type LIKE '%Lambda-GB-Second%' THEN 'Lambda GB x Sec.'
        WHEN line_item_usage_type LIKE '%Request%' THEN 'Lambda Requests'
        WHEN line_item_usage_type LIKE '%In-Bytes%' THEN 'Data Transfer (IN)'
        WHEN line_item_usage_type LIKE '%Out-Bytes%' THEN 'Data Transfer (Out)'
        WHEN line_item_usage_type LIKE '%Regional-Bytes%' THEN 'Data Transfer (Regional)'
        ELSE 'Other'
      END AS case_line_item_usage_type,
      line_item_resource_id,
      pricing_term,
      SUM(CAST(line_item_usage_amount AS DOUBLE)) AS sum_line_item_usage_amount,
      SUM(CAST(line_item_unblended_cost AS DECIMAL(16,8))) AS sum_line_item_unblended_cost, 
      SUM(CASE
        WHEN line_item_line_item_type = 'SavingsPlanCoveredUsage' THEN savings_plan_savings_plan_effective_cost
        WHEN line_item_line_item_type = 'SavingsPlanRecurringFee' THEN savings_plan_total_commitment_to_date - savings_plan_used_commitment
        WHEN line_item_line_item_type = 'SavingsPlanNegation' THEN 0 
        WHEN line_item_line_item_type = 'SavingsPlanUpfrontFee' THEN 0
        WHEN line_item_line_item_type = 'DiscountedUsage' THEN reservation_effective_cost
        WHEN line_item_line_item_type = 'RIFee' THEN reservation_unused_amortized_upfront_fee_for_billing_period + reservation_unused_recurring_fee
        WHEN line_item_line_item_type = 'Fee' AND reservation_reservation_a_r_n <> '' THEN 0
        ELSE line_item_unblended_cost 
      END) AS sum_amortized_cost
    FROM cur_db.data
      WHERE ${date_filter}
      AND product['product_name'] = 'AWS Lambda'
      AND line_item_line_item_type LIKE '%Usage%'
      AND product_product_family IN ('Data Transfer', 'Serverless')
      AND line_item_line_item_type  IN ('DiscountedUsage', 'Usage', 'SavingsPlanCoveredUsage')
    GROUP BY
      bill_payer_account_id,
      line_item_usage_account_id,
      line_item_line_item_type,
      DATE_FORMAT((line_item_usage_start_date),'%Y-%m-%d'),
      product['region'],
      6, -- refers to case_line_item_usage_type
      line_item_resource_id,
      pricing_term
  )

  UNION

  (
    SELECT
      bill_payer_account_id,
      line_item_usage_account_id,
      line_item_line_item_type,
      DATE_FORMAT((line_item_usage_start_date),'%Y-%m-%d') AS day_line_item_usage_start_date,
      product['region'],
      CASE
        WHEN line_item_usage_type LIKE '%Lambda-Edge-GB-Second%' THEN 'Lambda EDGE GB x Sec.'
        WHEN line_item_usage_type LIKE '%Lambda-Edge-Request%' THEN 'Lambda EDGE Requests'
        WHEN line_item_usage_type LIKE '%Lambda-GB-Second%' THEN 'Lambda GB x Sec.'
        WHEN line_item_usage_type LIKE '%Request%' THEN 'Lambda Requests'
        WHEN line_item_usage_type LIKE '%In-Bytes%' THEN 'Data Transfer (IN)'
        WHEN line_item_usage_type LIKE '%Out-Bytes%' THEN 'Data Transfer (Out)'
        WHEN line_item_usage_type LIKE '%Regional-Bytes%' THEN 'Data Transfer (Regional)'
        ELSE 'Other'
      END AS case_line_item_usage_type,
      line_item_resource_id,
      savings_plan_offering_type,
      SUM(CAST(line_item_usage_amount AS DOUBLE)) AS sum_line_item_usage_amount,
      SUM(CAST(savings_plan_savings_plan_effective_cost AS DECIMAL(16,8))) AS sum_savings_plan_savings_plan_effective_cost,
      SUM(CASE
        WHEN line_item_line_item_type = 'SavingsPlanCoveredUsage' THEN savings_plan_savings_plan_effective_cost
        WHEN line_item_line_item_type = 'SavingsPlanRecurringFee' THEN savings_plan_total_commitment_to_date - savings_plan_used_commitment
        WHEN line_item_line_item_type = 'SavingsPlanNegation' THEN 0
        WHEN line_item_line_item_type = 'SavingsPlanUpfrontFee' THEN 0
        WHEN line_item_line_item_type = 'DiscountedUsage' THEN reservation_effective_cost
        WHEN line_item_line_item_type = 'RIFee' THEN reservation_unused_amortized_upfront_fee_for_billing_period + reservation_unused_recurring_fee
        WHEN line_item_line_item_type = 'Fee' AND reservation_reservation_a_r_n <> '' THEN 0
        ELSE line_item_unblended_cost 
      END) AS sum_amortized_cost
    FROM 
      cur_db.data
    WHERE 
      line_item_usage_start_date >= date_add('month', -6, current_date)
      AND product['product_name'] = 'AWS Lambda'
      AND product_product_family IN ('Data Transfer', 'Serverless')
      AND line_item_line_item_type  IN ('DiscountedUsage', 'Usage', 'SavingsPlanCoveredUsage')
    GROUP BY
      bill_payer_account_id,
      line_item_usage_account_id,
      line_item_line_item_type,
      DATE_FORMAT((line_item_usage_start_date),'%Y-%m-%d'),
      product['region'],
      6, --refers to case_line_item_usage_type
      line_item_resource_id,
      savings_plan_offering_type
  )
) AS aggregatedTable

ORDER BY
  day_line_item_usage_start_date,
  sum_line_item_usage_amount,
  Sum_line_item_unblended_cost;
```


The query calculates both unblended and amortized costs, providing a clear view of where Lambda expenses are concentrated. This helps identify areas for optimization, such as reducing execution time or minimizing unnecessary data transfer.



###### Figure 7.9 AWS Lambda cost and usage breakdown


### Graviton-Based Instance Utilization

AWS Graviton instances offer better price-performance ratios, making them a preferred option for many workloads (they delivery up to 40% improvement over comparable current gen x86 processors). Graviton-based EC2 instances are available, and many other AWS services such as Amazon Relational Database Service, Amazon ElastiCache, Amazon EMR, and Amazon OpenSearch also support Graviton-based instance types.

This query provides insights into the usage and cost of Graviton-based instances for the last 6 months. Output is grouped by day, payer account ID, linked account ID, service, instance type, and region.


#### Listing 7.5 Query calculating cost and usage of Graviton based EC2 instances 


```
SELECT 
  DATE_TRUNC('day',line_item_usage_start_date) AS day_line_item_usage_start_date,
  bill_payer_account_id,
  line_item_usage_account_id,
  line_item_product_code,
  product_instance_type,
  product_region_code,
  SUM(CASE
    WHEN line_item_line_item_type = 'SavingsPlanCoveredUsage' THEN savings_plan_savings_plan_effective_cost
    WHEN line_item_line_item_type = 'DiscountedUsage' THEN reservation_effective_cost
    WHEN line_item_line_item_type = 'Usage' THEN line_item_unblended_cost
    ELSE 0 
  END) AS sum_amortized_cost, 
  SUM(line_item_usage_amount) as sum_line_item_usage_amount, 
  COUNT(DISTINCT(line_item_resource_id)) AS count_line_item_resource_id
FROM 
  cur_db.data
WHERE 
  line_item_usage_start_date >= date_add('month', -6, current_date)
  AND REGEXP_LIKE(line_item_usage_type, '.?[a-z]([1-9]|[1-9][0-9]).?.?[g][a-zA-Z]?\.')
  AND line_item_usage_type NOT LIKE '%EBSOptimized%' 
  AND (line_item_line_item_type = 'Usage'
    OR line_item_line_item_type = 'SavingsPlanCoveredUsage'
    OR line_item_line_item_type = 'DiscountedUsage'
  )
GROUP BY
  DATE_TRUNC('day',line_item_usage_start_date),
  bill_payer_account_id,
  line_item_usage_account_id,
  line_item_product_code,
  line_item_usage_type,
  product_instance_type,
  product_region_code
ORDER BY 
  day_line_item_usage_start_date DESC,
  sum_amortized_cost DESC;
```



### Spot Instance Cost Savings

Spot Instances offer significant cost savings, but tracking these savings requires comparing Spot prices to On-Demand prices. This query calculates average savings (over the past three months) for Spot Instances compared to the On-Demand public pricing for EC2 instances.


#### Listing 7.6 Query calculating monthly savings of Spot instances usage compared to On-demand instances


```
SELECT 
  product_instance_type, 
  product_region_code, 
  product['availability_zone'] as product_availability_zone,
  line_item_line_item_type, 
  SPLIT(billing_period, '-') [ 2 ] as split_month,
  -- SUM(line_item_unblended_cost) as "Spot Cost",  -- uncomment to reveal spot costs
  -- SUM(pricing_public_on_demand_cost) as "On-Demand", -- uncomment to reveal OD costs
  CAST(ROUND((AVG(1 - (line_item_unblended_cost / pricing_public_on_demand_cost)) * 100),2) as VARCHAR) || '%' as "avg_percentage_savings"
FROM 
  cur_db.data
WHERE 
  line_item_usage_type LIKE '%SpotUsage%'
  AND line_item_product_code = 'AmazonEC2'
  AND line_item_line_item_type = 'Usage'
  AND (CAST(CONCAT(billing_period, '-01') AS date) >= (DATE_TRUNC('month', current_date) - INTERVAL '1' MONTH)) -- lookback period
  -- AND product_instance_type = 'm7g.2xlarge' -- uncomment for specific instance type
GROUP BY 
  1, 2, 3, 4, 5
ORDER BY 
  product_instance_type,
  product['availability_zone']
```


This is a small subset of helpful queries you can perform to identify cost-saving opportunities and understand your cloud costs and usage. For a complete list of SQL queries, check the GitHub repository associated with the book, where you can find advanced queries to identify and optimize your AWS, GCP, and Azure compute resources.

While running SQL queries may be suitable for technical users, non-technical stakeholders involved in the FinOps process often require visual representations and interactive dashboards to gain insights effectively. That’s why you can leverage techniques covered in previous chapters to use FinOps dashboards built with QuickSight or take advantage of the out-of-the-box dashboards provided by the CUDOS Framework.
