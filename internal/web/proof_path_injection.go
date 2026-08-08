package web

// This file exists on this branch only. It is the near miss the code scanning
// check in #25 has to refuse: a file name taken straight out of a request and
// handed to the file system, which is the shape the volunteer surface and the
// subject ingest are both made of. It is deliberately ordinary code. It
// compiles, it is formatted, and the toolchain's own analyser and the linter
// both pass it, so a red run here is the analyser and nothing else.

import (
	"net/http"
	"os"
)

// ServeSubjectFile writes back the file the request names. The name reaches the
// file system unchecked, so a request asking for a path outside the collection
// is served exactly like one asking for a subject.
func ServeSubjectFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("file")
	body, err := os.ReadFile(name)
	if err != nil {
		http.Error(w, "no such subject", http.StatusNotFound)
		return
	}
	_, _ = w.Write(body)
}
