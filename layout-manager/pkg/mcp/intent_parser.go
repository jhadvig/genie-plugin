package mcp

import (
	"regexp"
	"strconv"
	"strings"
)

// IntentParser handles natural language parsing for widget operations
type IntentParser struct {
	// Command pattern matchers
	removePatterns   []*regexp.Regexp
	resizePatterns   []*regexp.Regexp
	movePatterns     []*regexp.Regexp
	addPatterns      []*regexp.Regexp
	updatePatterns   []*regexp.Regexp

	// Size patterns
	sizePatterns     []*regexp.Regexp

	// Position patterns
	positionPatterns []*regexp.Regexp
}

// NewIntentParser creates a new intent parser with compiled regex patterns
func NewIntentParser() *IntentParser {
	parser := &IntentParser{}
	parser.compilePatterns()
	return parser
}

// compilePatterns compiles all regex patterns for intent recognition
func (p *IntentParser) compilePatterns() {
	// Remove operation patterns
	p.removePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(remove|delete|get rid of|take away)\b`),
	}

	// Resize operation patterns
	p.resizePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(make (larger|bigger|smaller)|resize|expand|shrink)\b`),
		regexp.MustCompile(`(?i)\bresize.*?to\s+(\d+)x(\d+)\b`),
	}

	// Move operation patterns
	p.movePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(move|relocate|put|place)\b`),
		regexp.MustCompile(`(?i)\b(to|in|at)\s+(top|bottom)?\s*(left|right|center)\b`),
	}

	// Add operation patterns
	p.addPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(add|create|insert|place)\b`),
	}

	// Update/configure patterns
	p.updatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(change|update|modify|set|configure)\b`),
	}

	// Size extraction patterns
	p.sizePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\d+)x(\d+)`),                           // "4x3", "6x4"
		regexp.MustCompile(`(?i)\b(large|big|huge)\b`),                  // Size adjectives
		regexp.MustCompile(`(?i)\b(small|tiny|compact)\b`),
		regexp.MustCompile(`(?i)\b(medium|normal|standard)\b`),
		regexp.MustCompile(`(?i)\bmake.*?(larger|bigger|smaller)\b`),    // "make larger"
	}

	// Position extraction patterns
	p.positionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(top|bottom)\s+(left|right)\b`),       // "top left", "bottom right"
		regexp.MustCompile(`(?i)\b(left|right)\s+(side|corner)\b`),      // "left side", "right corner"
		regexp.MustCompile(`(?i)\bnext to\s+(?:the\s+)?(.+?)\b`),        // "next to the chart"
		regexp.MustCompile(`(?i)\b(above|below|left of|right of)\s+(?:the\s+)?(.+?)\b`), // "above the table"
	}
}

// ParseCommand parses a natural language command into an Intent
func (p *IntentParser) ParseCommand(command string) *Intent {
	command = strings.TrimSpace(strings.ToLower(command))

	intent := &Intent{
		Params: make(map[string]interface{}),
	}

	// Determine primary action
	intent.Action = p.parseAction(command)

	// Parse action-specific parameters
	switch intent.Action {
	case "resize":
		intent.SizeParams = p.parseSizeParams(command)
	case "move":
		intent.PositionParams = p.parsePositionParams(command)
	case "update":
		intent.PropsParams = p.parsePropsParams(command)
	}

	return intent
}

// parseAction determines the primary action from the command
func (p *IntentParser) parseAction(command string) string {
	// Check patterns in order of specificity
	for _, pattern := range p.removePatterns {
		if pattern.MatchString(command) {
			return "remove"
		}
	}

	for _, pattern := range p.resizePatterns {
		if pattern.MatchString(command) {
			return "resize"
		}
	}

	for _, pattern := range p.movePatterns {
		if pattern.MatchString(command) {
			return "move"
		}
	}

	for _, pattern := range p.addPatterns {
		if pattern.MatchString(command) {
			return "add"
		}
	}

	for _, pattern := range p.updatePatterns {
		if pattern.MatchString(command) {
			return "update"
		}
	}

	return "unknown"
}

// parseSizeParams extracts size parameters from resize commands
func (p *IntentParser) parseSizeParams(command string) *SizeParams {
	params := &SizeParams{}

	// Check for explicit dimensions (e.g., "resize to 6x4")
	dimensionPattern := regexp.MustCompile(`(?i)(\d+)x(\d+)`)
	if matches := dimensionPattern.FindStringSubmatch(command); len(matches) >= 3 {
		if w, err := strconv.Atoi(matches[1]); err == nil {
			params.Width = &w
		}
		if h, err := strconv.Atoi(matches[2]); err == nil {
			params.Height = &h
		}
		params.Mode = "absolute"
		return params
	}

	// Check for relative size changes
	if strings.Contains(command, "larger") || strings.Contains(command, "bigger") {
		delta := 1
		params.Delta = &delta
		params.Mode = "larger"
	} else if strings.Contains(command, "smaller") {
		delta := -1
		params.Delta = &delta
		params.Mode = "smaller"
	}

	// Check for size adjectives
	if regexp.MustCompile(`(?i)\b(large|big|huge)\b`).MatchString(command) {
		w, h := 6, 4
		params.Width = &w
		params.Height = &h
		params.Mode = "absolute"
	} else if regexp.MustCompile(`(?i)\b(small|tiny|compact)\b`).MatchString(command) {
		w, h := 2, 2
		params.Width = &w
		params.Height = &h
		params.Mode = "absolute"
	} else if regexp.MustCompile(`(?i)\b(medium|normal|standard)\b`).MatchString(command) {
		w, h := 4, 3
		params.Width = &w
		params.Height = &h
		params.Mode = "absolute"
	}

	return params
}

// parsePositionParams extracts position parameters from move commands
func (p *IntentParser) parsePositionParams(command string) *PositionParams {
	params := &PositionParams{}

	// Check for zone-based positioning (e.g., "top left", "bottom right")
	zonePattern := regexp.MustCompile(`(?i)\b(top|bottom)\s+(left|right)\b`)
	if matches := zonePattern.FindStringSubmatch(command); len(matches) >= 3 {
		params.Zone = matches[1] + "-" + matches[2]
		return params
	}

	// Check for simple directional commands (e.g., "to the left", "to the right")
	simpleDirectionPattern := regexp.MustCompile(`(?i)\bto\s+the\s+(left|right|top|bottom)\b`)
	if matches := simpleDirectionPattern.FindStringSubmatch(command); len(matches) >= 2 {
		params.Direction = matches[1]
		return params
	}

	// Check for basic directional commands (e.g., "left", "right")
	basicDirectionPattern := regexp.MustCompile(`(?i)\b(left|right|up|down)\b`)
	if matches := basicDirectionPattern.FindStringSubmatch(command); len(matches) >= 2 {
		direction := matches[1]
		if direction == "up" {
			direction = "top"
		} else if direction == "down" {
			direction = "bottom"
		}
		params.Direction = direction
		return params
	}

	// Check for relative positioning (e.g., "next to the chart")
	nextToPattern := regexp.MustCompile(`(?i)\bnext to\s+(?:the\s+)?(.+?)(?:\s|$)`)
	if matches := nextToPattern.FindStringSubmatch(command); len(matches) >= 2 {
		params.RelativeTo = strings.TrimSpace(matches[1])
		params.Direction = "right" // Default to right side
		return params
	}

	// Check for directional positioning (e.g., "above the table")
	directionPattern := regexp.MustCompile(`(?i)\b(above|below|left of|right of)\s+(?:the\s+)?(.+?)(?:\s|$)`)
	if matches := directionPattern.FindStringSubmatch(command); len(matches) >= 3 {
		params.RelativeTo = strings.TrimSpace(matches[2])
		direction := matches[1]
		switch direction {
		case "above":
			params.Direction = "above"
		case "below":
			params.Direction = "below"
		case "left of":
			params.Direction = "left"
		case "right of":
			params.Direction = "right"
		}
		return params
	}

	return params
}

// parsePropsParams extracts property update parameters
func (p *IntentParser) parsePropsParams(command string) map[string]interface{} {
	props := make(map[string]interface{})

	// Look for title updates (e.g., "change title to 'Sales Data'")
	titlePattern := regexp.MustCompile(`(?i)\btitle\s+to\s+['""](.+?)['""]`)
	if matches := titlePattern.FindStringSubmatch(command); len(matches) >= 2 {
		props["title"] = matches[1]
	}

	// Look for data source updates (e.g., "change data source to '/api/customers'")
	dataSourcePattern := regexp.MustCompile(`(?i)\bdata\s*source\s+to\s+['""](.+?)['""]`)
	if matches := dataSourcePattern.FindStringSubmatch(command); len(matches) >= 2 {
		props["dataSource"] = matches[1]
	}

	return props
}

// ParseWidgetSelector parses a widget selector description into matching criteria
func (p *IntentParser) ParseWidgetSelector(selector string) *WidgetMatcher {
	selector = strings.TrimSpace(strings.ToLower(selector))
	matcher := &WidgetMatcher{}

	// Extract component type
	componentTypes := []string{"chart", "table", "metric", "text", "image", "iframe"}
	for _, componentType := range componentTypes {
		if strings.Contains(selector, componentType) {
			matcher.ComponentType = componentType
			break
		}
	}

	// Extract title hints
	titlePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(sales|revenue|customer|order|product)\b`),
		regexp.MustCompile(`(?i)\bwith\s+(.+?)\s+(data|info|information)\b`),
		regexp.MustCompile(`(?i)\bshowing\s+(.+?)(?:\s|$)`),
	}

	for _, pattern := range titlePatterns {
		if matches := pattern.FindStringSubmatch(selector); len(matches) >= 2 {
			matcher.TitleContains = matches[1]
			break
		}
	}

	// Extract position hints
	positionHints := map[string]string{
		"top left":     "top-left",
		"top right":    "top-right",
		"bottom left":  "bottom-left",
		"bottom right": "bottom-right",
		"left side":    "left",
		"right side":   "right",
	}

	for hint, zone := range positionHints {
		if strings.Contains(selector, hint) {
			matcher.PositionZone = zone
			break
		}
	}

	// Extract size hints
	if strings.Contains(selector, "large") || strings.Contains(selector, "big") {
		matcher.SizeRange = &SizeRange{MinWidth: intPtr(5), MinHeight: intPtr(4)}
	} else if strings.Contains(selector, "small") || strings.Contains(selector, "tiny") {
		matcher.SizeRange = &SizeRange{MaxWidth: intPtr(3), MaxHeight: intPtr(2)}
	}

	return matcher
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}