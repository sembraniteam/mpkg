package config

import "fmt"

// ValidationError describes a failed validation for a specific module field.
type ValidationError struct {
	Module string
	Field  string
	Msg    string
}

func (e *ValidationError) Error() string {
	if e.Module == "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Msg)
	}
	return fmt.Sprintf("module %q: %s %s", e.Module, e.Field, e.Msg)
}
