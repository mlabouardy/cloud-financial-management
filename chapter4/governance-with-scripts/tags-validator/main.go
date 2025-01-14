package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-lambda-go/lambda"
)

func validateTags(tags map[string]string) []string {
	requiredTags := []string{"Environment", "Owner", "Project"}
	validEnvironments := map[string]bool{
		"Production":  true,
		"Staging":     true,
		"Development": true,
	}

	var errors []string

	for _, tag := range requiredTags {
		if _, exists := tags[tag]; !exists {
			errors = append(errors, fmt.Sprintf("Missing required tag: %s", tag))
		}
	}

	if env, exists := tags["Environment"]; exists {
		if !validEnvironments[env] {
			errors = append(errors, fmt.Sprintf("Invalid Environment value: %s", env))
		}
	}

	if owner, exists := tags["Owner"]; exists {
		emailRegex := `^[^@]+@[^@]+\.[^@]+$`
		matched, _ := regexp.MatchString(emailRegex, owner)
		if !matched {
			errors = append(errors, fmt.Sprintf("Invalid Owner format: %s", owner))
		}
	}

	if project, exists := tags["Project"]; exists {
		projectRegex := `^[a-zA-Z0-9-]+$`
		matched, _ := regexp.MatchString(projectRegex, project)
		if !matched {
			errors = append(errors, fmt.Sprintf("Invalid Project format: %s", project))
		}
	}

	return errors
}

func handler(event map[string]map[string]string) (map[string]interface{}, error) {
	tags, exists := event["tags"]
	if !exists {
		return nil, errors.New("Missing 'tags' in event")
	}

	validationErrors := validateTags(tags)

	if len(validationErrors) > 0 {
		return map[string]interface{}{
			"isValid": false,
			"errors":  validationErrors,
		}, nil
	}

	return map[string]interface{}{
		"isValid": true,
	}, nil
}

func main() {
	lambda.Start(handler)
}
