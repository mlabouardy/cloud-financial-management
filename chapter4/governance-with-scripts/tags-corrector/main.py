import boto3

def lambda_handler(event, context):
    ec2 = boto3.client('ec2')
    
    response = ec2.describe_instances() #A
    
    for reservation in response['Reservations']:
        for instance in reservation['Instances']:
            instance_id = instance['InstanceId']
            tags = instance.get('Tags', [])
            
            corrected_tags = []
            for tag in tags:
                key = tag['Key']
                value = tag['Value']
                
                corrected_key = key.capitalize() #B
                corrected_value = value.capitalize()
                
                if key != corrected_key or value != corrected_value:
                    corrected_tags.append({
                        'Key': corrected_key,
                        'Value': corrected_value
                    })
            
            if corrected_tags: #C
                ec2.create_tags(
                    Resources=[instance_id],
                    Tags=corrected_tags
                )
                print(f"Corrected tags for instance {instance_id}")

    return {
        'statusCode': 200,
        'body': 'Tag correction completed'
    }
