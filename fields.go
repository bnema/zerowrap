package zerowrap

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Common field name constants for consistent logging across applications.
const (
	FieldComponent = "component"
	FieldRequestID = "request_id"
	FieldTraceID   = "trace_id"
	FieldSpanID    = "span_id"
	FieldUserID    = "user_id"
	FieldError     = "error"
	FieldDuration  = "duration_ms"
)

// FromCtxWithField returns a logger with one additional field.
func FromCtxWithField(ctx context.Context, key string, value any) Logger {
	return Logger{addToContext(FromCtx(ctx).With(), key, value).Logger()}
}

// FromCtxWithFields returns a logger with multiple additional fields.
func FromCtxWithFields(ctx context.Context, fields map[string]any) Logger {
	log := FromCtx(ctx)
	c := log.With()
	for k, v := range fields {
		c = addToContext(c, k, v)
	}
	return Logger{c.Logger()}
}

// FromCtxWithStruct returns a logger with fields extracted from struct tags.
// Uses `log` struct tag for field names, falls back to `json` tag, then field name.
// Fields tagged with "-" are skipped.
func FromCtxWithStruct(ctx context.Context, s any) Logger {
	fields := extractFields(s)
	return FromCtxWithFields(ctx, fields)
}

// CtxWithField returns a new context with an enriched logger containing the field.
func CtxWithField(ctx context.Context, key string, value any) context.Context {
	log := FromCtxWithField(ctx, key, value)
	return WithCtx(ctx, log)
}

// CtxWithFields returns a new context with an enriched logger containing the fields.
func CtxWithFields(ctx context.Context, fields map[string]any) context.Context {
	log := FromCtxWithFields(ctx, fields)
	return WithCtx(ctx, log)
}

// CtxWithStruct returns a new context with an enriched logger containing fields from struct.
func CtxWithStruct(ctx context.Context, s any) context.Context {
	log := FromCtxWithStruct(ctx, s)
	return WithCtx(ctx, log)
}

// addToContext adds a field to the zerolog Context with type-specific methods for efficiency.
func addToContext(c zerolog.Context, key string, val any) zerolog.Context {
	switch v := val.(type) {
	case string:
		return c.Str(key, v)
	case int:
		return c.Int(key, v)
	case int8:
		return c.Int8(key, v)
	case int16:
		return c.Int16(key, v)
	case int32:
		return c.Int32(key, v)
	case int64:
		return c.Int64(key, v)
	case uint:
		return c.Uint(key, v)
	case uint8:
		return c.Uint8(key, v)
	case uint16:
		return c.Uint16(key, v)
	case uint32:
		return c.Uint32(key, v)
	case uint64:
		return c.Uint64(key, v)
	case float32:
		return c.Float32(key, v)
	case float64:
		return c.Float64(key, v)
	case bool:
		return c.Bool(key, v)
	case error:
		return c.AnErr(key, v)
	case time.Time:
		return c.Time(key, v)
	case time.Duration:
		return c.Dur(key, v)
	case []byte:
		return c.Bytes(key, v)
	case []string:
		return c.Strs(key, v)
	default:
		return c.Interface(key, v)
	}
}

// extractFields extracts loggable fields from a struct using reflection.
// Priority: `log` tag > `json` tag > field name (lowercased with underscores).
func extractFields(s any) map[string]any {
	fields := make(map[string]any)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fields
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fields
	}

	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Determine field name from tags
		name := field.Tag.Get("log")
		if name == "" {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				name = strings.Split(jsonTag, ",")[0]
			}
		}
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		// Skip fields tagged with "-"
		if name == "-" {
			continue
		}

		fieldVal := v.Field(i)

		// Skip zero values for pointers and interfaces
		if (fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface) && fieldVal.IsNil() {
			continue
		}

		fields[name] = fieldVal.Interface()
	}

	return fields
}

// toSnakeCase converts PascalCase/camelCase to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r + 32) // lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
