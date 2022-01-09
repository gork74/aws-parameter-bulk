package server

import (
	"net/http"
	"strings"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

// return 404 not found if index.html do not exist when passing url with path suffix /
func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s != nil && s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := nfs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})

	dynamicMiddleware := alice.New(app.session.LoadAndSave, noSurf)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Post("/", dynamicMiddleware.ThenFunc(app.postHome))
	mux.Post("/reset", dynamicMiddleware.ThenFunc(app.postReset))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
