package prometheus

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Guardrail name constants for use with ParseGuardrails
const (
	GuardrailDisallowExplicitNameLabel = "disallow-explicit-name-label"
	GuardrailRequireLabelMatcher       = "require-label-matcher"
	GuardrailDisallowBlanketRegex      = "disallow-blanket-regex"
	GuardrailMaxMetricCardinality      = "max-metric-cardinality"
	GuardrailMaxLabelCardinality       = "max-label-cardinality"
)

// Guardrails provides safety checks for PromQL queries based on configurable rules.
type Guardrails struct {
	// DisallowExplicitNameLabel prevents queries using explicit {__name__="..."} syntax
	DisallowExplicitNameLabel bool
	// RequireLabelMatcher ensures all vector selectors have at least one non-name label matcher
	RequireLabelMatcher bool
	// DisallowBlanketRegex prevents expensive regex patterns like .* or .+ on any label
	DisallowBlanketRegex bool
	// MaxMetricCardinality sets the maximum allowed series count per metric (0 = disabled)
	MaxMetricCardinality uint64
	// MaxLabelCardinality sets the maximum allowed label value count for blanket regex
	// (0 = always disallow regex matcher provided DisallowBlanketRegex is true)
	MaxLabelCardinality uint64
}

// DefaultGuardrails returns a Guardrails instance with all safety checks enabled.
func DefaultGuardrails() *Guardrails {
	return &Guardrails{
		DisallowExplicitNameLabel: true,
		RequireLabelMatcher:       true,
		DisallowBlanketRegex:      true,
		MaxMetricCardinality:      20000,
		MaxLabelCardinality:       500,
	}
}

func ParseGuardrails(value string) (*Guardrails, error) {
	value = strings.TrimSpace(value)

	switch strings.ToLower(value) {
	case "none":
		return nil, nil
	case "all", "":
		return DefaultGuardrails(), nil
	}

	g := &Guardrails{}
	names := strings.Split(value, ",")
	for _, name := range names {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}

		switch name {
		case GuardrailDisallowExplicitNameLabel:
			g.DisallowExplicitNameLabel = true
		case GuardrailRequireLabelMatcher:
			g.RequireLabelMatcher = true
		case GuardrailDisallowBlanketRegex:
			g.DisallowBlanketRegex = true
		default:
			return nil, fmt.Errorf("unknown guardrail: %q (valid options: %s, %s, %s, %s, %s)",
				name, GuardrailDisallowExplicitNameLabel, GuardrailRequireLabelMatcher,
				GuardrailDisallowBlanketRegex, GuardrailMaxMetricCardinality, GuardrailMaxLabelCardinality)
		}
	}

	return g, nil
}

// IsSafeQuery analyzes a PromQL query string and returns false if it's
// deemed unsafe or too expensive based on the configured rules.
// If client is provided and MaxMetricCardinality is set, it checks TSDB metric cardinality.
// If client is provided and MaxLabelCardinality is set, it checks TSDB label cardinality for blanket regex.
//
// Returns (false, error) only if the query syntax is invalid or TSDB check fails.
// Returns (false, nil) if the query is valid but violates a rule.
// Returns (true, nil) if the query is valid and passes all rules.
func (g *Guardrails) IsSafeQuery(ctx context.Context, query string, client v1.API) (bool, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return false, fmt.Errorf("failed to parse query: %w", err)
	}

	foundUnsafe := false
	hasBlanketRegex := false

	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		switch n := node.(type) {
		case *parser.VectorSelector:
			hasNonNameMatcher := false

			for _, m := range n.LabelMatchers {
				// Rule 1: Check for explicit __name__ label query
				if g.DisallowExplicitNameLabel {
					if m.Name == labels.MetricName && n.Name == "" {
						foundUnsafe = true
						return fmt.Errorf("unsafe")
					}
				}

				if m.Name != labels.MetricName {
					hasNonNameMatcher = true
				}

				// Rule 3: Check for expensive regex matchers on *any* label i.e blanket matchers
				if g.DisallowBlanketRegex {
					isRegex := m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp
					if isRegex && (m.Value == ".*" || m.Value == ".+") {
						// If MaxLabelCardinality is set, defer the check to TSDB lookup
						if g.MaxLabelCardinality > 0 {
							hasBlanketRegex = true
						} else {
							// MaxLabelCardinality is 0, always disallow blanket regex here
							foundUnsafe = true
							return fmt.Errorf("unsafe")
						}
					}
				}
			}

			// Rule 2: All vector selectors must have at least one non-name label matcher
			if g.RequireLabelMatcher && !hasNonNameMatcher {
				foundUnsafe = true
				return fmt.Errorf("unsafe")
			}
		}
		return nil
	})

	if foundUnsafe {
		return false, nil
	}

	// If blanket regex was found but we can't check TSDB (no client), reject the query
	// this is only set to true when MaxLabelCardinality is set and DisallowBlanketRegex is true
	if hasBlanketRegex && (client == nil || ctx == nil) {
		return false, nil
	}

	// Rule 4: Check metric cardinality if enabled and client is provided
	if g.MaxMetricCardinality > 0 && client != nil && ctx != nil {
		metricNames, err := ExtractMetricNames(query)
		if err != nil {
			return false, fmt.Errorf("failed to extract metric names: %w", err)
		}

		if len(metricNames) > 0 {
			tsdbResult, err := client.TSDB(ctx)
			if err != nil {
				return false, fmt.Errorf("failed to get TSDB stats: %w", err)
			}

			seriesCountByMetric := make(map[string]uint64)
			for _, stat := range tsdbResult.SeriesCountByMetricName {
				seriesCountByMetric[stat.Name] = stat.Value
			}

			for _, metricName := range metricNames {
				if count, exists := seriesCountByMetric[metricName]; exists {
					if count > g.MaxMetricCardinality {
						return false, nil
					}
				}
			}
		}
	}

	// Rule 5: Check label cardinality for blanket regex if enabled and client is provided
	if hasBlanketRegex && g.MaxLabelCardinality > 0 && client != nil && ctx != nil {
		labelNames, err := ExtractBlanketRegexLabels(query)
		if err != nil {
			return false, fmt.Errorf("failed to extract blanket regex labels: %w", err)
		}

		if len(labelNames) > 0 {
			// TODO: Cache this result.
			tsdbResult, err := client.TSDB(ctx)
			if err != nil {
				return false, fmt.Errorf("failed to get TSDB stats: %w", err)
			}

			labelValueCountByLabel := make(map[string]uint64)
			for _, stat := range tsdbResult.LabelValueCountByLabelName {
				labelValueCountByLabel[stat.Name] = stat.Value
			}

			for _, labelName := range labelNames {
				if count, exists := labelValueCountByLabel[labelName]; exists {
					if count > g.MaxLabelCardinality {
						return false, nil
					}
				}
			}
		}
	}

	return true, nil
}

func ExtractMetricNames(query string) ([]string, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	metricNames := make(map[string]bool)
	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if vs, ok := node.(*parser.VectorSelector); ok {
			if vs.Name != "" {
				metricNames[vs.Name] = true
			}
			// Also check for __name__ label matchers
			for _, m := range vs.LabelMatchers {
				if m.Name == labels.MetricName && m.Type == labels.MatchEqual {
					metricNames[m.Value] = true
				}
			}
		}
		return nil
	})

	result := make([]string, 0, len(metricNames))
	for name := range metricNames {
		result = append(result, name)
	}
	return result, nil
}

// ExtractBlanketRegexLabels extracts label names that use blanket regex patterns (.* or .+).
func ExtractBlanketRegexLabels(query string) ([]string, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	labelNames := make(map[string]bool)
	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if vs, ok := node.(*parser.VectorSelector); ok {
			for _, m := range vs.LabelMatchers {
				isRegex := m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp
				if isRegex && (m.Value == ".*" || m.Value == ".+") {
					labelNames[m.Name] = true
				}
			}
		}
		return nil
	})

	result := make([]string, 0, len(labelNames))
	for name := range labelNames {
		result = append(result, name)
	}
	return result, nil
}
