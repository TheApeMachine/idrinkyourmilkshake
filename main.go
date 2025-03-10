package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/openai"
)

func main() {
	log.Info("Starting application")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	log.Info("Initializing OpenAI client")
	client := openai.NewClient(apiKey)

	log.Info("Creating background context")
	ctx := context.Background()
	client = client.WithContext(ctx)

	log.Info("Creating conversation buffer with system and user prompts")
	buffer := openai.NewBuffer(
		`
		You are an advanced API integration expert.
		You work with a specialized API Integration Engine that relies on a configuration file to drive all parts of the integration.
		You will be given a URL to a page of API documentation and your job is to extract the API endpoints and data models from the 
		documentation and generate a configuration object.
		You have access to a full Chrome browser as a tool, so you can navigate the documentation and do whatever is needed to extract the information.
		You also have access to an HTTP request tool, so you can interact with APIs when needed, provided you have the correct authentication details.
		You are placed in an iteration loop, so you have the ability to work for as long as needed to complete the task.
		Do not return the configuration object until you have finished the task, and don't return an "example" configuration object, or provide
		any information that is not real, or not part of the documentation.
		Use the tools at your disposal to complete the task, then to finish, return the configuration object.
		Most likely your best strategy is to navigate the documentation, and use a smart approach to extract information.
		The HTTP tool is only really useful if you have been provided with the authentication details, but its there if you need it.

		The most likely scenario is that you just need to get the page content, and you should have all the information you need.

		Here is an example of the configuration object you will return:

		{
			"integration": "dyflexis",
			"base_url": "https://api.dyflexis.com",
			"auth": {
				"type": "basic",
				"outputs": [
					{
						"key": "username",
						"value": "admin"
					},
					{
						"key": "password",
						"value": "password"
					}
				]
			},
			"jobs": [
				{
					"id": "get_employees",
					"steps": [
						{
							"type": "http",
							"id": "login",
							"endpoint": "/login",
							"method": "POST",
							"body": [{
								"key": "username",
								"value": "{{auth.username}}"
							}],
							"outputs": {
								"token": "{{response.body.token}}"
							}
						},	
						{
							"type": "http",
							"id": "get_employees",
							"endpoint": "/employees",
							"method": "GET",
							"headers": {
								"x-tamigo-token": "{{login.outputs.token}}"
							},
							"outputs": {
								"employees": "{{response.body.employees}}"
							}
						},
						{
							"type": "mongodb",
							"id": "store_employees",
							"collection": "User",
							"operation": "upsert",
							"documents": [
								{
									"key": "username",
									"value": "{{get_employees.outputs.employees}}"
								},
								{
									"key": "email",
									"value": "{{get_employees.outputs.employees.email}}"
								}
							]
						}
					]
				}
			]
		}
		`,
		`
		Here is the documentation URL for the API: https://developer.dyflexis.com/v3
		`,
	)

	log.Info("Starting OpenAI client execution with max iterations", "maxIterations", 20)
	result, err := client.Execute(buffer, 20)
	if err != nil {
		log.Fatal("Error executing OpenAI client", "error", err)
	}

	fmt.Println(result)
}
