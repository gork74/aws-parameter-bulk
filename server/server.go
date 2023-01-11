package server

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/gork74/aws-parameter-bulk/pkg/util"
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type application struct {
	logger        *zerolog.Logger
	session       *scs.SessionManager
	templateCache map[string]*template.Template
	ssmClient     *util.AWSSSM
}

func ListenAndServe(logger *zerolog.Logger, address string) {

	var err error

	InitTemplates()
	session := scs.New()
	session.Lifetime = 24 * 30 * time.Hour

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Fatal().Msgf("Template cache Error %s", err)
	}

	ssmClient := util.NewSSM()
	app := &application{logger, session, templateCache, ssmClient}

	srv := &http.Server{
		Addr:    address,
		Handler: app.routes(),
	}

	rand.Seed(time.Now().UnixNano())

	addrString := address
	if strings.HasPrefix(addrString, ":") {
		addrString = fmt.Sprintf("localhost%s", address)
	}
	app.logger.Info().Msgf("Starting server on http://%s", addrString)
	err = srv.ListenAndServe()
	if err != nil {
		app.logger.Fatal().Err(err).Msg("Startup failed")
	}
}
