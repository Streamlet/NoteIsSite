package server

import (
	"github.com/Streamlet/NoteIsSite/note"
	"net/http"
	"os"
)

func newRouter(noteRoot string, templateRoot string) (http.Handler, error) {
	notesRouter, err := note.NewRouter(noteRoot, templateRoot)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := notesRouter.Route(r.RequestURI)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write(content)
	})

	return mux, nil
}
