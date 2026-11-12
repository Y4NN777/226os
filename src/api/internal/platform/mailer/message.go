package mailer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

// build renders the RFC 5322 message. Dot-stuffing is left to net/smtp's Data writer.
func build(from *mail.Address, to []*mail.Address, m Message, now time.Time, rng io.Reader, msgIDHost string) ([]byte, error) {
	id := make([]byte, 16)
	if _, err := io.ReadFull(rng, id); err != nil {
		return nil, fmt.Errorf("mailer: generate message id: %w", err)
	}
	rcpt := make([]string, len(to))
	for i, a := range to {
		rcpt[i] = a.String()
	}
	var b bytes.Buffer // writes to a bytes.Buffer cannot fail
	fmt.Fprintf(&b, "From: %s\r\n", from.String())
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(rcpt, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", encodeSubject(m.Subject))
	fmt.Fprintf(&b, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", hex.EncodeToString(id), msgIDHost)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qp := quotedprintable.NewWriter(&b)
	_, _ = qp.Write([]byte(m.Text)) // bytes.Buffer cannot fail
	_ = qp.Close()                  // bytes.Buffer cannot fail
	return b.Bytes(), nil
}

// maxWordPayload keeps each encoded word (12 bytes of framing) within 68 characters, so a header
// line stays under the 78-character recommendation of RFC 5322.
const maxWordPayload = 56

// encodeSubject returns s unchanged when it is ASCII and cannot be read as an RFC 2047 encoded
// word (no "=?"), otherwise RFC 2047 Q-encoded words folded
// onto continuation lines. mime.QEncoding would emit one unfolded line for a long subject.
func encodeSubject(s string) string {
	ascii := !strings.Contains(s, "=?")
	for i := 0; i < len(s); i++ {
		ascii = ascii && s[i] < utf8.RuneSelf
	}
	if ascii {
		return s
	}
	var words []string
	var cur strings.Builder
	flush := func() {
		words = append(words, "=?utf-8?q?"+cur.String()+"?=")
		cur.Reset()
	}
	for _, r := range s {
		enc := qEncodeRune(r)
		if cur.Len()+len(enc) > maxWordPayload {
			flush()
		}
		cur.WriteString(enc)
	}
	flush()
	return strings.Join(words, "\r\n ")
}

// qEncodeRune encodes r for an RFC 2047 Q word: letters and digits as is, space as "_", all else as =XX bytes.
func qEncodeRune(r rune) string {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return string(r)
	case r == ' ':
		return "_"
	}
	var b strings.Builder
	for _, c := range []byte(string(r)) {
		fmt.Fprintf(&b, "=%02X", c)
	}
	return b.String()
}
