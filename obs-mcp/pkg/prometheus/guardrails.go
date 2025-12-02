package prometheus

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Guardrail name constants for use with ParseGuardrails
const (
	GuardrailDisallowExplicitNameLabel = "disallow-explicit-name-label"
	GuardrailRequireLabelMatcher       = "require-label-matcher"
	GuardrailDisallowBlanketRegex      = "disallow-blanket-regex"
)

// Guardrails provides safety checks for PromQL queries based on configurable rules.
type Guardrails struct {
	// DisallowExplicitNameLabel prevents queries using explicit {__name__="..."} syntax
	DisallowExplicitNameLabel bool
	// RequireLabelMatcher ensures all vector selectors have at least one non-name label matcher
	RequireLabelMatcher bool
	// DisallowBlanketRegex prevents expensive regex patterns like .* or .+ on any label
	DisallowBlanketRegex bool
}

// DefaultGuardrails returns a Guardrails instance with all safety checks enabled.
func DefaultGuardrails() *Guardrails {
	return &Guardrails{
		DisallowExplicitNameLabel: true,
		RequireLabelMatcher:       true,
		DisallowBlanketRegex:      true,
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
			return nil, fmt.Errorf("unknown guardrail: %q (valid options: %s, %s, %s)",
				name, GuardrailDisallowExplicitNameLabel, GuardrailRequireLabelMatcher, GuardrailDisallowBlanketRegex)
		}
	}

	return g, nil
}

// IsSafeQuery analyzes a PromQL query string and returns false if it's
// deemed unsafe or too expensive based on the configured rules.
//
// Returns (false, error) only if the query syntax is invalid.
// Returns (false, nil) if the query is valid but violates a rule.
// Returns (true, nil) if the query is valid and passes all rules.
func (g *Guardrails) IsSafeQuery(query string) (bool, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return false, fmt.Errorf("failed to parse query: %w", err)
	}

	foundUnsafe := false
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
					if isRegex {
						if m.Value == ".*" || m.Value == ".+" {
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

	return true, nil
}
