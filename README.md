# I Drink Your Milkshake 🥤

> _"I drink your API documentation milkshake. I drink it up!"_

[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![OpenAI](https://img.shields.io/badge/OpenAI-API-412991?style=flat&logo=openai)](https://openai.com/)
[![GoRod](https://img.shields.io/badge/go--rod-Browser_Automation-00ADD8?style=flat)](https://github.com/go-rod/rod)

## 🧠 AI-Powered API Integration Assistant

**I Drink Your Milkshake** is an intelligent tool that automates the tedious process of integrating with third-party APIs. Using the power of OpenAI's GPT-4, it analyzes API documentation, extracts endpoints and data models, and generates the configuration needed to drive your integration.

No more spending hours manually reading through API docs and building integration configs by hand!

## ✨ Features

- 🤖 **AI-Powered Analysis**: Leverages OpenAI's advanced models to understand API documentation
- 🌐 **Browser Automation**: Uses Go-Rod to navigate and interact with API documentation sites
- 🔍 **Smart Extraction**: Intelligently identifies endpoints, parameters, and data models
- 🔄 **HTTP Request Testing**: Can make test requests to verify API understanding
- 📝 **Configuration Generation**: Outputs a structured configuration file ready for your integration engine

## 💻 How It Works

1. The application starts an OpenAI session with a specialized prompt that turns GPT-4 into an API integration expert
2. You provide a URL to the API documentation
3. The AI navigates through the documentation using a real Chrome browser
4. It extracts essential information about endpoints and data models
5. When needed, it can make HTTP requests to test and verify its understanding
6. Finally, it generates a comprehensive configuration object representing the API

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
