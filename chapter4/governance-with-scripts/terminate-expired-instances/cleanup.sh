#!/bin/bash

# Get current date in YYYY-MM-DD format
current_date=$(date +%Y-%m-%d)

# List instances with DecommissionDate tag
instances=$(aws ec2 describe-instances --filters "Name=tag-key,Values=DecommissionDate" --query "Reservations[].Instances[].{ID:InstanceId,DecommissionDate:Tags[?Key=='DecommissionDate'].Value|[0]}" --output text)

while read -r instance_id decommission_date; do
    if [[ "$decommission_date" < "$current_date" || "$decommission_date" == "$current_date" ]]; then
        echo "Terminating instance $instance_id (Decommission Date: $decommission_date)"
        aws ec2 terminate-instances --instance-ids "$instance_id"
    else
        echo "Instance $instance_id not yet due for decommissioning (Decommission Date: $decommission_date)"
    fi
done <<< "$instances"