import re

def validate_tags(tags):
    required_tags = ['Environment', 'Owner', 'Project']
    valid_environments = ['Production', 'Staging', 'Development']
    
    errors = []
    
    for tag in required_tags: #A
        if tag not in tags:
            errors.append(f"Missing required tag: {tag}")
    
    if 'Environment' in tags: #B
        if tags['Environment'] not in valid_environments:
            errors.append(f"Invalid Environment value: {tags['Environment']}")
    
    if 'Owner' in tags: #C
        if not re.match(r"[^@]+@[^@]+\.[^@]+", tags['Owner']):
            errors.append(f"Invalid Owner format: {tags['Owner']}")
    
    if 'Project' in tags: #D
        if not re.match(r"^[a-zA-Z0-9-]+$", tags['Project']):
            errors.append(f"Invalid Project format: {tags['Project']}")
    
    return errors

def lambda_handler(event, context):
    tags = event['tags']
    validation_errors = validate_tags(tags)
    
    if validation_errors:
        return {
            'isValid': False,
            'errors': validation_errors
        }
    else:
        return {
            'isValid': True
        }
