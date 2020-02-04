package server

import (
	"github.com/Streamlet/NoteIsSite/note"
	"log"
	"mime"
	"net/http"
	"net/url"
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
				log.Println(r.RequestURI, "404:", err.Error())
				w.WriteHeader(http.StatusNotFound)
			} else {
				log.Println(r.RequestURI, "500:", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			if urlParts, err := url.Parse(r.RequestURI); err == nil {
				path := urlParts.EscapedPath()
				for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
					if path[i] == '.' {
						if mimeType := mime.TypeByExtension(path[i:]); mimeType != "" {
							w.Header().Add("Content-Type", mimeType)
						}
						break
					}
				}
			}
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write(content)
	})

	return mux, nil
}
