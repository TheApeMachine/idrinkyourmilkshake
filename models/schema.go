package models

// Auth represents authentication details for an API
type Auth struct {
	Type    string   `json:"type"`
	Outputs []Output `json:"outputs"`
}

// Headers represents HTTP headers for API requests
type Headers struct {
	Key   string `json:"key" jsonschema:"description=Name of the header,required"`
	Value string `json:"value" jsonschema:"description=Value of the header,required"`
}

// Input represents input data for an API request
type Input struct {
	Headers []Header `json:"headers" jsonschema:"description=HTTP headers for the request,required"`
	Body    Body     `json:"body" jsonschema:"description=Body of the request,required"`
}

type Header struct {
	Name  string `json:"name" jsonschema:"description=Name of the header,required"`
	Value string `json:"value" jsonschema:"description=Value of the header,required"`
}

type Body struct {
	Key   string `json:"key" jsonschema:"description=Key of the body,required"`
	Value string `json:"value" jsonschema:"description=Value of the body,required"`
}

// Output represents output data from an API request
type Output struct {
	Key   string `json:"key" jsonschema:"description=Key of the output,required"`
	Value string `json:"value" jsonschema:"description=Value of the output,required"`
}

// Step represents a step in a job
type Step struct {
	Type       string     `json:"type" jsonschema:"description=Type of step to execute,required,enum=http,enum=mongodb"`
	ID         string     `json:"id" jsonschema:"description=Unique identifier for the step,required"`
	Endpoint   string     `json:"endpoint,omitempty" jsonschema:"description=API endpoint to call"`
	Method     string     `json:"method,omitempty" jsonschema:"description=HTTP method to use"`
	Outputs    []Output   `json:"outputs,omitempty" jsonschema:"description=Output data from the step"`
	Collection string     `json:"collection,omitempty" jsonschema:"description=Collection name for database operations"`
	Operation  string     `json:"operation,omitempty" jsonschema:"description=Operation to perform on the collection"`
	Documents  []Document `json:"documents,omitempty" jsonschema:"description=Documents to insert or update in the collection"`
}

type Document struct {
	Key   string `json:"key" jsonschema:"description=Key of the document,required"`
	Value string `json:"value" jsonschema:"description=Value of the document,required"`
}

// Job represents a job with steps
type Job struct {
	ID    string `json:"id" jsonschema:"description=Unique identifier for the job,required"`
	Steps []Step `json:"steps" jsonschema:"description=Steps to execute in the job,required"`
}

// APIConfig represents the complete API configuration
type APIConfig struct {
	Integration string `json:"integration" jsonschema:"description=The name of the integration,required"`
	BaseURL     string `json:"base_url" jsonschema:"description=The base URL,required"`
	Auth        Auth   `json:"auth" jsonschema:"description=The authentication details,required"`
	Jobs        []Job  `json:"jobs" jsonschema:"description=The jobs to run,required"`
}
