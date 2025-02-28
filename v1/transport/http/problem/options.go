package problem

// Option represents a functional option to modify a Problem.
type Option func(*Problem)

// WithType returns an option to set a custom problem type.
func WithType(t string) Option {
	return func(p *Problem) {
		p.Type = t
	}
}

// WithField returns an option to add a single extension field.
func WithField(key string, value any) Option {
	return func(p *Problem) {
		if p.Fields == nil {
			p.Fields = make(Fields)
		}
		p.Fields[key] = value
	}
}

// WithFields returns an option to add multiple extension fields at once.
// It accepts a map of fields, for example using log.Fields from logrus.
func WithFields(fields Fields) Option {
	return func(p *Problem) {
		if p.Fields == nil {
			p.Fields = make(Fields)
		}
		for key, value := range fields {
			p.Fields[key] = value
		}
	}
}

type Fields map[string]any
