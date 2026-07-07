package utils

import (
	"fmt"
	"io"
	"net/http"
)

// DefaultMaxNotifyBodySize caps the bytes read from an inbound WeChat callback
// (notify) request body. Real WeChat callbacks are well under 100 KiB; the
// 1 MiB cap gives ~10× headroom while preventing an unauthenticated remote
// caller from exhausting process memory with a huge or chunked body.
const DefaultMaxNotifyBodySize int64 = 1 << 20

// ReadNotifyBody reads up to limit bytes from r.Body and returns them. A body
// that would exceed limit is rejected with an explicit "notify body exceeds N
// bytes" error rather than being silently truncated. limit<=0 uses
// DefaultMaxNotifyBodySize.
//
// It is the single shared implementation behind every inbound notify handler
// (offiaccount / oplatform / work-wechat ISV / merchant), so the size cap
// cannot drift between call sites. Callers MUST invoke it before parsing
// (xml/json Unmarshal) so a hostile body is bounded before it can drive
// unbounded allocation, and MUST still verify the signature over the returned
// bytes — ReadNotifyBody only bounds the read, it does not authenticate.
//
// It does not close r.Body; the caller retains ownership (typically via
// defer r.Body.Close()), matching net/http handler conventions and the
// existing merchant notify handler.
func ReadNotifyBody(r *http.Request, limit int64) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("nil *http.Request")
	}
	if r.Body == nil {
		return nil, fmt.Errorf("nil request body")
	}
	if limit <= 0 {
		limit = DefaultMaxNotifyBodySize
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read notify body: %w", err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("notify body exceeds %d bytes", limit)
	}
	return body, nil
}
