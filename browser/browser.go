package browser

import (
	"fmt"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/log"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/theapemachine/idrinkyourmilkshake/models"
)

// Global page instance to be reused across browser operations
var page *rod.Page
var browser *rod.Browser

func init() {
	log.Info("Initializing browser")
	u := launcher.New().
		Set("user-data-dir", "path").
		Delete("--headless").
		MustLaunch()

	browser = rod.New().ControlURL(u).MustConnect()
	page = browser.MustPage("")
	log.Info("Browser initialized successfully")
}

type BrowserExtractor struct {
	ToolName        string           `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string           `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  models.Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string         `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

func NewBrowserExtractor() models.ToolType {
	return &BrowserExtractor{
		ToolName:        "extract_page_content",
		ToolDescription: "Extracts the full content from the current page and converts it to markdown",
		ToolParameters: models.Parameter{
			Type:       "object",
			Properties: []models.Property{},
			Required:   true,
		},
		Required: []string{},
	}
}

func (be *BrowserExtractor) Name() string {
	return be.ToolName
}

func (be *BrowserExtractor) Description() string {
	return be.ToolDescription
}

func (be *BrowserExtractor) Execute(args map[string]any) (string, error) {
	selector := "body"

	log.Info("Extracting content with selector", "selector", selector)

	element := page.MustElement(selector)
	if element == nil {
		log.Error("Element not found", "selector", selector)
		return "", fmt.Errorf("element not found: %s", selector)
	}

	// Start highlighting the element - use blue for extraction
	highlightElement(page, selector)

	log.Info("Getting HTML content from element")
	html, err := element.HTML()
	if err != nil {
		log.Error("Error getting HTML", "error", err)
		return "", fmt.Errorf("error getting HTML: %w", err)
	}

	log.Info("Converting HTML to markdown", "htmlSize", len(html))
	markdown, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		log.Error("Error converting HTML to markdown", "error", err)
		return "", fmt.Errorf("error converting HTML to markdown: %w", err)
	}

	log.Info("Successfully extracted and converted content", "markdownSize", len(markdown))
	return markdown, nil
}

func (be *BrowserExtractor) Schema() any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	}
}

type BrowserNavigator struct {
	ToolName        string           `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string           `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  models.Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string         `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

func NewBrowserNavigator() models.ToolType {
	return &BrowserNavigator{
		ToolName:        "browser_navigate",
		ToolDescription: "Navigates to a URL",
		ToolParameters: models.Parameter{
			Type: "object",
			Properties: []models.Property{
				{
					Name:        "url",
					Type:        "string",
					Description: "The URL to navigate to",
				},
			},
			Required: true,
		},
		Required: []string{"url"},
	}
}

func (bn *BrowserNavigator) Name() string {
	return bn.ToolName
}

func (bn *BrowserNavigator) Description() string {
	return bn.ToolDescription
}

func (bn *BrowserNavigator) Execute(args map[string]any) (string, error) {
	var (
		url string
		ok  bool
	)

	if url, ok = args["url"].(string); !ok {
		log.Error("No URL provided")
		return "", fmt.Errorf("no URL provided")
	}

	log.Info("Navigating browser to URL", "url", url)
	page.MustNavigate(url).MustWaitStable()
	log.Info("Successfully navigated to URL and page is stable", "url", url)
	return "Successfully navigated to " + url, nil
}

func (bn *BrowserNavigator) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to navigate to",
			},
		},
		"required": []string{"url"},
	}
}

type BrowserClicker struct {
	ToolName        string           `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string           `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  models.Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string         `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

func NewBrowserClicker() models.ToolType {
	return &BrowserClicker{
		ToolName:        "browser_click",
		ToolDescription: "Clicks an element on the current page",
		ToolParameters: models.Parameter{
			Type: "object",
			Properties: []models.Property{
				{
					Name:        "selector",
					Type:        "string",
					Description: "The selector of the element to click",
				},
			},
			Required: true,
		},
		Required: []string{"selector"},
	}
}

func (bc *BrowserClicker) Name() string {
	return bc.ToolName
}

func (bc *BrowserClicker) Description() string {
	return bc.ToolDescription
}

func (bc *BrowserClicker) Execute(args map[string]any) (string, error) {
	var (
		selector string
		ok       bool
	)

	if selector, ok = args["selector"].(string); !ok {
		log.Error("No selector provided")
		return "", fmt.Errorf("no selector provided")
	}

	log.Info("Clicking element with selector", "selector", selector)

	// Start highlighting the element before clicking - use green for clicking
	highlightElement(page, selector)

	// Add a small delay to make the highlight visible before the click
	time.Sleep(500 * time.Millisecond)

	// Click the element
	element := page.MustElement(selector)
	element.MustClick()

	log.Info("Successfully clicked element", "selector", selector)
	return "clicked " + selector, nil
}

func (bc *BrowserClicker) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"selector": map[string]any{
				"type":        "string",
				"description": "The selector of the element to click",
			},
		},
		"required": []string{"selector"},
	}
}

type BrowserJavaScriptExecutor struct {
	ToolName        string           `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string           `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  models.Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string         `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

func NewBrowserJavaScriptExecutor() models.ToolType {
	return &BrowserJavaScriptExecutor{
		ToolName:        "browser_execute_js",
		ToolDescription: "Executes JavaScript in the browser",
		ToolParameters: models.Parameter{
			Type: "object",
			Properties: []models.Property{
				{
					Name:        "script",
					Type:        "string",
					Description: "The JavaScript code to execute, you must supply a function that returns a string, no exceptions",
				},
			},
			Required: true,
		},
		Required: []string{"script"},
	}
}

func (bje *BrowserJavaScriptExecutor) Name() string {
	return bje.ToolName
}

func (bje *BrowserJavaScriptExecutor) Description() string {
	return bje.ToolDescription
}

func (bje *BrowserJavaScriptExecutor) Execute(args map[string]any) (string, error) {
	var (
		script string
		ok     bool
	)

	if script, ok = args["script"].(string); !ok {
		log.Error("No script provided")
		return "", fmt.Errorf("no script provided")
	}

	log.Info("Executing JavaScript in browser", "scriptLength", len(script))
	out := page.MustEval(script).Str()
	log.Info("JavaScript execution successful", "outputLength", len(out))

	return out, nil
}

func (bje *BrowserJavaScriptExecutor) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"script": map[string]any{
				"type":        "string",
				"description": "The JavaScript code to execute, you must supply a function that returns a string, no exceptions",
			},
		},
		"required": []string{"script"},
	}
}

func ExtractPageContent(browser *rod.Browser, selector string) (string, error) {
	log.Info("Extracting page content with selector", "selector", selector)
	page := browser.MustPage("")
	content := page.MustElement("body").MustHTML()

	log.Info("Converting HTML to markdown", "htmlSize", len(content))
	markdown, err := htmltomarkdown.ConvertString(content)
	if err != nil {
		log.Error("Error converting HTML to markdown", "error", err)
		return "", fmt.Errorf("error converting HTML to markdown: %w", err)
	}

	log.Info("Successfully extracted and converted content", "markdownSize", len(markdown))
	return markdown, nil
}

// highlightElement creates a simple highlight around an element and scrolls to it.
// Returns a unique ID for the created highlight element, which can be used to remove it later.
func highlightElement(page *rod.Page, selector string) {
	page.MustEval(`(selector) => {
		const el = document.querySelector(selector);

		if (!el) {
			return "Element not found";
		}

		window.scrollTo({
			top: el.offsetTop - (window.innerHeight / 2),
			behavior: "smooth"
		});

		const rect = el.getBoundingClientRect();
		highlighter = document.createElement("div");
		document.body.appendChild(highlighter);
		
		highlighter.style.position = "absolute";
		highlighter.style.left = (rect.left + window.scrollX) + "px";
		highlighter.style.top = (rect.top + window.scrollY) + "px";
		highlighter.style.width = rect.width + "px";
		highlighter.style.height = rect.height + "px";
		highlighter.style.pointerEvents = "none";
		highlighter.style.zIndex = "10000";
		highlighter.style.backgroundColor = "rgba(250, 224, 95, 0.1)";

		// Perform a double pulse animation to highlight the element, then remove it.
		highlighter.animate([
			{ backgroundColor: "rgba(250, 224, 95, 0.1)", offset: 0.3 },
			{ backgroundColor: "rgba(250, 224, 95, 0.9)", offset: 0.6 },
		], {
			duration: 1000,
			iterations: 2,
		});

		setTimeout(() => {
			highlighter.remove();
		}, 2000);

		return ""
	}`, selector)
}
