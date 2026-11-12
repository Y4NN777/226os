package idempotency

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// envelopeVersion is the "v" of the response_body envelope.
const envelopeVersion = 1

// maxEnvelopeValue bounds a stored header value or content type so that the body-less
// envelope always fits the column (Location is at most 512 characters in the contract).
const maxEnvelopeValue = 2048

// replayedHeaders is the allowlist of response headers kept and replayed (besides
// Content-Type): never Set-Cookie, never the X-Request-Id of the first request.
var replayedHeaders = [...]string{"Location", "ETag"}

// envelope is what response_body holds: the table has no column for the content type or
// the headers, and the column is a JSON object anyway.
type envelope struct {
	V           int               `json:"v"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        json.RawMessage   `json:"body,omitempty"`
	BodyOmitted bool              `json:"body_omitted,omitempty"`
}

// encodeEnvelope renders resp as an envelope. The body is left out, with body_omitted, when
// withBody is false or the response has none; the allowlisted headers are kept only when
// withHeaders is true.
func encodeEnvelope(resp Response, withBody, withHeaders bool) ([]byte, error) {
	e := envelope{V: envelopeVersion}
	if len(resp.ContentType) <= maxEnvelopeValue {
		e.ContentType = resp.ContentType
	}
	for _, name := range replayedHeaders {
		if v := resp.Headers[name]; withHeaders && v != "" && len(v) <= maxEnvelopeValue {
			if e.Headers == nil {
				e.Headers = map[string]string{}
			}
			e.Headers[name] = v
		}
	}
	if withBody && isObject(resp.Body) {
		e.Body = resp.Body
	} else {
		e.BodyOmitted = true
	}
	out, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("encode idempotency envelope: %w", err)
	}
	return out, nil
}

// encodeEnvelopes renders the envelopes to try, richest first, each distinct from the one
// before: the full envelope, the body-less one that keeps the headers, and the minimal one
// (version, content type, body_omitted) for when PostgreSQL refuses a header value too.
func encodeEnvelopes(resp Response) ([][]byte, error) {
	var out [][]byte
	for _, tier := range [...]struct{ body, headers bool }{{true, true}, {false, true}, {false, false}} {
		b, err := encodeEnvelope(resp, tier.body, tier.headers)
		if err != nil {
			return nil, err
		}
		if len(out) == 0 || !bytes.Equal(out[len(out)-1], b) {
			out = append(out, b)
		}
	}
	return out, nil
}

// decodeEnvelope reads response_body back. Anything that is not a v1 envelope (a row of an
// older shape: any other object, or NULL) is the body of an application/json response.
func decodeEnvelope(status int, stored []byte) Response {
	var e envelope
	dec := json.NewDecoder(bytes.NewReader(stored))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil || e.V != envelopeVersion {
		resp := Response{Status: status, Body: stored}
		if stored != nil {
			resp.ContentType = "application/json"
		}
		return resp
	}
	return Response{Status: status, ContentType: e.ContentType, Headers: e.Headers, Body: e.Body}
}

// isObject reports whether b starts a JSON object; the column holds objects only, and the
// envelope keeps just those bodies.
func isObject(b []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeft(b, " \t\r\n"), []byte("{"))
}
