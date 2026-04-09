package config

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

// dotSeparator is the separator between section and field name in dot-notation keys.
const dotSeparator = "."

// maxSplitParts is the maximum number of parts when splitting a dot-notation key.
const maxSplitParts = 2

// tagSplitParts is the maximum number of parts when splitting a yaml tag value.
const tagSplitParts = 2

// GetValue returns the Go value for the given dot-notation key from cfg.
//
// Accepted key forms:
//   - "section.name"  — returns the scalar/slice value of a specific field.
//   - "section"       — returns a map[string]any of all fields in the section.
//
// Returns [clerrors.ErrOptionNotFound] for unknown keys.
func GetValue(cfg *ConfigData, key string) (any, error) {
	parts := strings.SplitN(key, dotSeparator, maxSplitParts)
	section := strings.ToLower(parts[0])

	// Section-only key: return all fields as a map.
	if len(parts) == 1 {
		sv, err := sectionStruct(cfg, section)
		if err != nil {
			return nil, err
		}
		return structToMap(sv), nil
	}

	fieldKey := parts[1]

	sv, err := sectionStruct(cfg, section)
	if err != nil {
		return nil, err
	}

	fv, err := fieldByYAMLTag(reflect.ValueOf(sv), fieldKey)
	if err != nil {
		return nil, unknownKeyError(key)
	}

	return fv.Interface(), nil
}

// SetValue parses rawValue and sets it on the field identified by the dot-notation key.
//
// Type coercions applied based on the target field kind:
//   - bool         — "true" / "false"
//   - int          — decimal integer string
//   - float64      — decimal float string
//   - []string     — comma-separated values (each entry trimmed)
//   - time.Duration — duration string parsed by [time.ParseDuration]
//   - string       — used as-is
//
// Returns [clerrors.ErrOptionNotFound] for unknown keys, or a descriptive error on
// type conversion failures.
func SetValue(cfg *ConfigData, key string, rawValue string) error {
	parts := strings.SplitN(key, dotSeparator, maxSplitParts)
	if len(parts) != maxSplitParts {
		return unknownKeyError(key)
	}

	section := strings.ToLower(parts[0])
	fieldKey := parts[1]

	sv, err := sectionStruct(cfg, section)
	if err != nil {
		return err
	}

	// sectionStruct returns a pointer so we get the addressable Elem for setting.
	rv := reflect.ValueOf(sv).Elem()

	fv, err := fieldByYAMLTag(rv, fieldKey)
	if err != nil {
		return unknownKeyError(key)
	}

	if !fv.CanSet() {
		return fmt.Errorf("%w: %s", clerrors.ErrFieldNotSettable, key)
	}

	return setFieldValue(fv, key, rawValue)
}

// AllKeys returns all valid dot-notation keys registered in the MetadataRegistry,
// sorted alphabetically.
func AllKeys() []string {
	reg := NewMetadataRegistry()
	opts := reg.GetAll()
	keys := make([]string, 0, len(opts))
	for _, opt := range opts {
		keys = append(keys, fmt.Sprintf("%s.%s", opt.Section, opt.Name))
	}
	sort.Strings(keys)
	return keys
}

// sectionStruct returns a pointer to the section struct within cfg.
// Returns [clerrors.ErrUnknownSection] for unrecognised section names.
func sectionStruct(cfg *ConfigData, section string) (any, error) {
	switch section {
	case SectionDefaults:
		return &cfg.Defaults, nil
	case SectionSearch:
		return &cfg.Search, nil
	case SectionOutput:
		return &cfg.Output, nil
	case SectionAPI:
		return &cfg.API, nil
	default:
		return nil, fmt.Errorf("%w: %s (available: defaults, search, output, api)",
			clerrors.ErrUnknownSection, section)
	}
}

// structToMap converts a pointer-to-struct into map[string]any keyed by yaml tags.
func structToMap(sv any) map[string]any {
	rv := reflect.ValueOf(sv)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	rt := rv.Type()
	result := make(map[string]any, rv.NumField())

	for i := range rv.NumField() {
		tag := yamlTagName(rt.Field(i))
		if tag == "" || tag == "-" {
			continue
		}
		result[tag] = rv.Field(i).Interface()
	}

	return result
}

// fieldByYAMLTag searches a struct value for the field whose yaml tag matches name.
// v may be a pointer or a struct Value; the function dereferences pointers automatically.
func fieldByYAMLTag(v reflect.Value, name string) (reflect.Value, error) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	t := v.Type()
	for i := range v.NumField() {
		tag := yamlTagName(t.Field(i))
		if tag == name {
			return v.Field(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("%w: yaml tag %q", clerrors.ErrFieldNotFound, name)
}

// yamlTagName extracts the base name from a yaml struct tag (before any comma).
func yamlTagName(f reflect.StructField) string {
	tag := f.Tag.Get("yaml")
	if tag == "" {
		return ""
	}
	return strings.SplitN(tag, ",", tagSplitParts)[0]
}

// setFieldValue converts rawValue to the kind required by fv and assigns it.
// Duration is handled as a special case before dispatching on reflect.Kind.
func setFieldValue(fv reflect.Value, key, rawValue string) error {
	// Handle time.Duration before the generic kind switch.
	if fv.Type() == reflect.TypeFor[time.Duration]() {
		return setDuration(fv, key, rawValue)
	}

	switch fv.Kind() { //nolint:exhaustive // remaining kinds are not valid config field types
	case reflect.String:
		fv.SetString(rawValue)

	case reflect.Bool:
		return setBool(fv, key, rawValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fv, key, rawValue)

	case reflect.Float32, reflect.Float64:
		return setFloat(fv, key, rawValue)

	case reflect.Slice:
		return setStringSlice(fv, key, rawValue)

	default:
		return fmt.Errorf("%w: %s (kind %s)", clerrors.ErrUnsupportedFieldKind, key, fv.Kind())
	}

	return nil
}

func setDuration(fv reflect.Value, key, rawValue string) error {
	d, err := time.ParseDuration(rawValue)
	if err != nil {
		return fmt.Errorf("cannot parse %q as duration for %q: %w", rawValue, key, err)
	}
	fv.Set(reflect.ValueOf(d))
	return nil
}

func setBool(fv reflect.Value, key, rawValue string) error {
	b, err := strconv.ParseBool(rawValue)
	if err != nil {
		return fmt.Errorf("cannot parse %q as bool for %q: %w", rawValue, key, err)
	}
	fv.SetBool(b)
	return nil
}

func setInt(fv reflect.Value, key, rawValue string) error {
	n, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as int for %q: %w", rawValue, key, err)
	}
	fv.SetInt(n)
	return nil
}

func setFloat(fv reflect.Value, key, rawValue string) error {
	f, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as float for %q: %w", rawValue, key, err)
	}
	fv.SetFloat(f)
	return nil
}

func setStringSlice(fv reflect.Value, key, rawValue string) error {
	if fv.Type().Elem().Kind() != reflect.String {
		return fmt.Errorf("%w: %s", clerrors.ErrUnsupportedSliceType, key)
	}
	parts := strings.Split(rawValue, ",")
	elems := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			elems = append(elems, trimmed)
		}
	}
	fv.Set(reflect.ValueOf(elems))
	return nil
}

// unknownKeyError constructs a helpful error listing all valid keys.
func unknownKeyError(key string) error {
	available := strings.Join(AllKeys(), ", ")
	return fmt.Errorf("%w: %s (available: %s)", clerrors.ErrOptionNotFound, key, available)
}
