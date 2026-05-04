package pay

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// requireIdempotencyKey is a defensive guard for money-moving endpoints
// (transfer, profitsharing) that accept a `body any` parameter. WeChat
// requires an idempotency key on these endpoints; if a caller forgets to
// populate it, WeChat returns an opaque error and the merchant may issue
// duplicate disbursements before the failure is noticed.
//
// This helper extracts the named field from body and rejects empty or
// missing values before the network round-trip. It accepts:
//   - map[string]any / map[string]string                            (looked up directly)
//   - struct or *struct with the JSON tag matching key               (extracted via tag)
//   - any other type that json.Marshal serialises to a JSON object   (extracted via marshal/unmarshal)
//
// The check is intentionally conservative: if body's structure cannot be
// inspected (e.g. it is a struct with no matching field, or marshals to
// a non-object), the helper allows the request through rather than
// blocking it — this preserves the SDK's "thin wrapper" philosophy for
// types it does not know about. The cost of a false negative (one extra
// network round-trip with a clearer WeChat error) is acceptable; the cost
// of a false positive (blocking a legitimate request) is not.
func requireIdempotencyKey(body any, key string) error {
	if body == nil {
		return fmt.Errorf("pay: body is required (must include %s)", key)
	}

	// Direct map paths.
	switch m := body.(type) {
	case map[string]any:
		return checkMapValue(m[key], key)
	case map[string]string:
		if v, ok := m[key]; ok {
			if v == "" {
				return fmt.Errorf("pay: %s must be non-empty", key)
			}
			return nil
		}
		return fmt.Errorf("pay: %s is required", key)
	}

	// Struct path: look for a field whose json tag matches key.
	v := reflect.ValueOf(body)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return fmt.Errorf("pay: body is nil (must include %s)", key)
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if found, val, ok := lookupJSONField(v, key); found {
			if !ok {
				return nil // field exists but unexported — let it through
			}
			if isEmpty(val) {
				return fmt.Errorf("pay: %s must be non-empty", key)
			}
			return nil
		}
		// Struct lacks a json tag for key; fall back to marshal probe so we
		// can still catch the case where the struct has a method-driven
		// MarshalJSON that produces the field.
	}

	// Marshal probe: serialise body to JSON, decode into map, look up key.
	raw, err := json.Marshal(body)
	if err != nil {
		// If we cannot marshal it, we cannot validate. Allow through; the
		// downstream postV3 will surface the marshal failure with context.
		return nil
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		// Body marshals to a non-object (e.g. an array). Cannot validate.
		return nil
	}
	return checkMapValue(probe[key], key)
}

func checkMapValue(v any, key string) error {
	if v == nil {
		return fmt.Errorf("pay: %s is required", key)
	}
	if s, ok := v.(string); ok && s == "" {
		return fmt.Errorf("pay: %s must be non-empty", key)
	}
	return nil
}

// lookupJSONField finds a struct field whose json tag matches key (the part
// before any comma). Returns (found, value, exported).
func lookupJSONField(v reflect.Value, key string) (bool, reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" {
			continue
		}
		// Strip ",omitempty" etc.
		name := tag
		for j := 0; j < len(tag); j++ {
			if tag[j] == ',' {
				name = tag[:j]
				break
			}
		}
		if name == key {
			return true, v.Field(i), f.IsExported()
		}
	}
	return false, reflect.Value{}, false
}

func isEmpty(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}
