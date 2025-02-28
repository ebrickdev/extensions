package problem

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// Problem represents an RFC7807-compliant error response.
type Problem struct {
	Type     string `json:"type"`     // A URI reference that identifies the problem type.
	Title    string `json:"title"`    // A short, human-readable summary of the problem.
	Status   int    `json:"status"`   // The HTTP status code.
	Detail   string `json:"detail"`   // A detailed explanation of the error.
	Instance string `json:"instance"` // A URI reference that identifies the specific occurrence of the error.
	Fields   Fields `json:"-"`        // Additional, custom fields.
}

// MarshalJSON customizes the JSON output to merge standard fields with any extensions.
// If Fields is empty, no extra keys are added.
func (p *Problem) MarshalJSON() ([]byte, error) {
	base := Fields{
		"type":     p.Type,
		"title":    p.Title,
		"status":   p.Status,
		"detail":   p.Detail,
		"instance": p.Instance,
	}
	// Merge extensions without overriding base fields.
	for k, v := range p.Fields {
		if _, exists := base[k]; !exists {
			base[k] = v
		}
	}
	return json.Marshal(base)
}

// New creates a new Problem instance using the given Gin context and applies any provided options.
func New(c *gin.Context, status int, title, detail string, opts ...Option) *Problem {
	prob := &Problem{
		Type:     "about:blank", // Default type; can be overridden with WithType.
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.URL.Path, // Default instance is the request URL path.
		Fields:   make(map[string]any),
	}
	for _, opt := range opts {
		opt(prob)
	}
	return prob
}

// AbortWithProblem creates a Problem and immediately aborts the request with a JSON response.
func AbortWithProblem(c *gin.Context, status int, title, detail string, opts ...Option) {
	p := New(c, status, title, detail, opts...)
	c.AbortWithStatusJSON(status, p)
}

// WriteProblem writes the Problem as a JSON response without aborting the context.
func WriteProblem(c *gin.Context, p *Problem) {
	c.JSON(p.Status, p)
}
