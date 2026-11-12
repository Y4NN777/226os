package cursor

import "encoding/base64"

// SignForTest signs an arbitrary body with the codec key, to exercise the
// parser on payloads that carry a valid signature but a bad structure.
func SignForTest(c *Codec, body []byte) string {
	return base64.RawURLEncoding.EncodeToString(append(body, c.sign(body)...))
}
