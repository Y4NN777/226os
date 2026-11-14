package httpx

import "net/http"

// ViewerScoped marks a response whose body depends on who asks (a viewer_vote or viewer_application
// field read from the session): a shared cache must key it by cookie and may not store it, and a private
// cache revalidates before reuse.
func ViewerScoped(w http.ResponseWriter) {
	h := w.Header()
	h.Add("Vary", "Cookie")
	h.Set("Cache-Control", "private, no-cache")
}
