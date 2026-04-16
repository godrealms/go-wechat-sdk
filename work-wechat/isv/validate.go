package isv

import "fmt"

// requireNonEmpty returns a validation error when value is empty. The returned
// error is formatted "isv: <method>: <field> is required", matching the style
// used across the rest of the SDK so callers can string-match on behavior if
// needed.
func requireNonEmpty(method, field, value string) error {
	if value == "" {
		return fmt.Errorf("isv: %s: %s is required", method, field)
	}
	return nil
}

// requirePositive returns a validation error when id is not > 0. Used for
// fields like agentid/tagid/deptid that are int in the request body but must
// be provided by the caller.
func requirePositive(method, field string, id int) error {
	if id <= 0 {
		return fmt.Errorf("isv: %s: %s must be > 0", method, field)
	}
	return nil
}

// validWWMediaTypes enumerates the WeCom (企业微信) temp-media upload types.
// Distinct from mini-program's {image,voice,video,thumb} — WeCom accepts
// "file" instead of "thumb".
var validWWMediaTypes = map[string]struct{}{
	"image": {}, "voice": {}, "video": {}, "file": {},
}
