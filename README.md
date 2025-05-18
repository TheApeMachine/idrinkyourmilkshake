# I Drink Your Milkshake 🥤

> _"I drink your API milkshake. I drink it up!"_

[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![OpenAI](https://img.shields.io/badge/OpenAI-API-412991?style=flat&logo=openai)](https://openai.com/)
[![GoRod](https://img.shields.io/badge/go--rod-Browser_Automation-00ADD8?style=flat)](https://github.com/go-rod/rod)

## 🧠 AI-Powered API Integration Assistant

**I Drink Your Milkshake** is an intelligent tool that automates the tedious process of integrating with third-party APIs. Using the power of OpenAI's GPT-4o, it analyzes API documentation, extracts endpoints and data models, and generates the configuration needed to drive your integration.

No more spending hours manually reading through API docs and building integration configs by hand!

## ✨ Features

- 🤖 **AI-Powered Analysis**: Leverages OpenAI's advanced models to understand API documentation
- 🌐 **Browser Automation**: Uses Go-Rod to navigate and interact with API documentation sites
- 🔍 **Smart Extraction**: Intelligently identifies endpoints, parameters, and data models
- 🔄 **HTTP Request Testing**: Can make test requests to verify API understanding
- 📝 **Configuration Generation**: Outputs a structured configuration file ready for your integration engine

## 💻 How It Works

There are two distinct, and decoupled processes that make up this project.

### API Discovery & Mapping

The purpose here is to end up with a configuration object that can drive the engine, once the discovery and mapping stage is complete.

The following things are the steps that need to take place at this stage.

1. The AI is given a couple of browser-based tools so it can go online and read the API documentation, from the URL provided on the command line. This is often just an HTML page, which will be converted to Markdown to limit the amount of tokens that would come with reading the raw HTML.
2. The AI also has a couple of tools to inspect the MongoDB database, which it can use to identify the target collections and schemas.
3. The AI is given a structured output jsonschema configuration, to make sure it is forced to structure its final output in a controlled manner, which will become the configuration object that will drive the engine.

### Engine

The engine is the mechanism that will actually run the integration, once the previous step is fully complete and we have a configuration object to drive the engine.

This will be deployed on Kubernetes as a cron job, and basically just make calls to the external API, and using the configuration object pull, convert, and load the API data into the correct MongoDB collections.

## 🚀 Getting Started

### Prerequisites

- Go 1.18+
- OpenAI API key
- Chrome/Chromium installed (for browser automation)

### Installation

```bash
# Clone the repository
git clone https://github.com/theapemachine/idrinkyourmilkshake.git

# Navigate to the project
cd idrinkyourmilkshake

# Install dependencies
go mod download
```

### Usage

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

Run the application:

```bash
go run main.go
```

You can also use the built-in CLI commands:

```bash
# Generate a configuration by analyzing API docs
idrinkyourmilkshake integrate -d https://example.com/api/docs -o integration-config.json

# Execute the integration using the generated configuration
idrinkyourmilkshake run -c integration-config.json
```

By default, the application will process the Dyflexis API documentation at [dyflexis](https://developer.dyflexis.com/v3).

To analyze a different API, modify the user prompt in `main.go`.

## 🔍 Under the Hood

This tool combines several powerful technologies:

- **OpenAI GPT-4o-mini**: For understanding API documentation and generating configurations
- **Go-Rod**: For browser automation and DOM manipulation
- **Charmbracelet Log**: For beautiful, structured logging
- **Tiktoken**: For token counting and context management

The application creates an execution loop where:

1. The model analyzes the current context and requests tools (browser navigation, content extraction, etc.)
2. The application executes these tools and feeds the results back to the model
3. This continues until the model has gathered enough information to generate the final API configuration

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues or pull requests.
