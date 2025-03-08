package models

// Auth represents authentication details for an API
type Auth struct {
	Type     string   `json:"type"`
	Endpoint string   `json:"endpoint"`
	Method   string   `json:"method"`
	Inputs   []Input  `json:"inputs"`
	Outputs  []Output `json:"outputs"`
}

// Map represents a data mapping configuration
type Map struct {
	ExternalID string `json:"ExternalId" jsonschema:"description=External identifier for the entity,required"`
	Name       string `json:"Name" jsonschema:"description=Name of the entity,required"`
	Email      string `json:"Email" jsonschema:"description=Email address of the entity,required"`
}

// Headers represents HTTP headers for API requests
type Headers struct {
	XTamigoToken string `json:"x-tamigo-token" jsonschema:"description=Authentication token for Tamigo API,required"`
}

// Input represents input data for an API request
type Input struct {
	Headers map[string]string `json:"headers" jsonschema:"description=HTTP headers for the request,required"`
	Body    map[string]any    `json:"body" jsonschema:"description=Body of the request,required"`
}

// Output represents output data from an API request
type Output struct {
	Employees string `json:"employees" jsonschema:"description=JSON string containing employee data"`
}

// Step represents a step in a job
type Step struct {
	Type       string `json:"type" jsonschema:"description=Type of step to execute,required"`
	Name       string `json:"name" jsonschema:"description=Name of the step,required"`
	Endpoint   string `json:"endpoint,omitempty" jsonschema:"description=API endpoint to call"`
	Method     string `json:"method,omitempty" jsonschema:"description=HTTP method to use"`
	Inputs     Input  `json:"inputs,omitempty" jsonschema:"description=Input data for the step"`
	Outputs    Output `json:"outputs,omitempty" jsonschema:"description=Output data from the step"`
	Input      string `json:"input,omitempty" jsonschema:"description=Input reference for the step"`
	Map        Map    `json:"map,omitempty" jsonschema:"description=Mapping configuration for data transformation"`
	Collection string `json:"collection,omitempty" jsonschema:"description=Collection name for database operations"`
	Operation  string `json:"operation,omitempty" jsonschema:"description=Operation to perform on the collection"`
	MatchField string `json:"match_field,omitempty" jsonschema:"description=Field to match when performing operations"`
}

// Job represents a job with steps
type Job struct {
	Name  string `json:"name" jsonschema:"description=Name of the job,required"`
	Steps []Step `json:"steps" jsonschema:"description=Steps to execute in the job,required"`
}

// APIConfig represents the complete API configuration
type APIConfig struct {
	Integration string `json:"integration" jsonschema:"description=The name of the integration,required"`
	AccountID   string `json:"account_id" jsonschema:"description=The account ID,required"`
	BaseURL     string `json:"base_url" jsonschema:"description=The base URL,required"`
	Auth        Auth   `json:"auth" jsonschema:"description=The authentication details,required"`
	Jobs        []Job  `json:"jobs" jsonschema:"description=The jobs to run,required"`
}
