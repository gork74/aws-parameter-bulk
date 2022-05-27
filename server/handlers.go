package server

import (
	"github.com/gork74/aws-parameter-bulk/pkg/forms"
	"github.com/gork74/aws-parameter-bulk/pkg/models"
	"github.com/gork74/aws-parameter-bulk/pkg/util"
	"net/http"
	"strings"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	view := &templateData{}
	if app.session.Get(r.Context(), "view") == nil {
		app.logger.Info().Msg("view is nil")
		view = &templateData{
			Form:       forms.New(nil),
			NamesLeft:  "",
			Different:  false,
			NamesRight: "",
			Compare:    make([]models.ValueCompare, 0),
		}
		app.session.Put(r.Context(), "view", view)
	} else {
		app.logger.Info().Msg("loading view")
		viewSess := app.session.Get(r.Context(), "view")
		view = viewSess.(*templateData)
		view.Compare = updateCompares(view.Compare)
	}
	app.render(w, r, "home.page.tmpl", view)
}

func (app *application) postReset(w http.ResponseWriter, r *http.Request) {
	app.logger.Info().Msg("resetting view")
	view := &templateData{
		Form:       forms.New(nil),
		NamesLeft:  "",
		JsonLeft:   false,
		Different:  false,
		NamesRight: "",
		JsonRight:  false,
		Compare:    make([]models.ValueCompare, 0),
	}
	app.session.Put(r.Context(), "view", view)
	app.render(w, r, "home.page.tmpl", view)
}

func (app *application) postHome(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.logger.Error().Msgf("Error parsing form: %s", err)
		return
	}
	jsonLeft := false
	jsonRight := false
	recursiveLeft := false
	recursiveRight := false

	form := forms.New(r.PostForm)
	form.Required("namesleft")
	namesLeft := strings.TrimSpace(form.Get("namesleft"))
	namesRight := strings.TrimSpace(form.Get("namesright"))
	jsonLeftFlag := form.Get("jsonleft")
	recursiveLeftFlag := form.Get("recursiveleft")
	jsonRightFlag := form.Get("jsonright")
	recursiveRightFlag := form.Get("recursiveright")
	if jsonLeftFlag == "on" {
		jsonLeft = true
	}
	if recursiveLeftFlag == "on" {
		recursiveLeft = true
	}
	if jsonRightFlag == "on" {
		jsonRight = true
	}
	if recursiveRightFlag == "on" {
		recursiveRight = true
	}
	flagsLeft := util.Flags{
		false,
		jsonLeft,
		false,
		false,
		false,
		false,
		recursiveLeft,
		false,
		false,
	}
	flagsRight := util.Flags{
		false,
		jsonRight,
		false,
		false,
		false,
		false,
		recursiveRight,
		false,
		false,
	}

	app.logger.Debug().Msgf("Namesright: '%s'", namesRight)
	app.logger.Debug().Msgf("JsonLeft: '%s'", jsonLeftFlag)
	app.logger.Debug().Msgf("JsonRight: '%s'", jsonRightFlag)

	// If there are any errors, redisplay the form.
	if !form.Valid() {
		app.logger.Error().Msgf("Names not valid")
		if err != nil {
			app.session.Put(r.Context(), "flasherror", "Error reading template data")
			app.render(w, r, "error.page.tmpl", &templateData{})
			return
		}
		app.session.Put(r.Context(), "flasherror", "Names not valid")

		app.render(w, r, "home.page.tmpl", &templateData{Form: form})
		return
	}

	resultLeft, err := app.ssmClient.GetParams(&namesLeft, flagsLeft)
	if err != nil {
		app.logger.Error().Msg(err.Error())
		app.session.Put(r.Context(), "flasherror", "Error reading values from the left side input: "+err.Error())
		app.render(w, r, "error.page.tmpl", &templateData{})
		return
	}

	sortedLeftNames := util.GetSortedNamesFromParams(resultLeft)

	compares := make([]models.ValueCompare, 0)

	for _, name := range sortedLeftNames {
		value := resultLeft[name]
		compare := models.ValueCompare{
			LeftName:     name,
			LeftOriginal: value,
			LeftValue:    value,
			Different:    false,
		}
		compares = append(compares, compare)
	}

	if namesRight != "" {
		resultRight, err := app.ssmClient.GetParams(&namesRight, flagsRight)
		if err != nil {
			app.logger.Error().Msg(err.Error())
			app.session.Put(r.Context(), "flasherror", "Error reading values from the right side input: "+err.Error())
			app.render(w, r, "error.page.tmpl", &templateData{})
			return
		}

		sortedRightNames := util.GetSortedNamesFromParams(resultRight)

		index := 0
		for _, name := range sortedRightNames {
			value := resultRight[name]
			compare := models.ValueCompare{
				RightName:     name,
				RightOriginal: value,
				RightValue:    value,
				Different:     false,
			}
			if index < len(compares) {
				compares[index].RightName = name
				compares[index].RightOriginal = value
				compares[index].RightValue = value
			} else {
				compares = append(compares, compare)
			}
			index++
		}
	}

	view := &templateData{
		Form:       form,
		NamesLeft:  namesLeft,
		JsonLeft:   jsonLeft,
		NamesRight: namesRight,
		JsonRight:  jsonRight,
		Compare:    compares,
	}
	view.Compare = updateCompares(view.Compare)
	app.session.Put(r.Context(), "view", view)
	app.render(w, r, "home.page.tmpl", view)
	return
}

func updateCompares(compares []models.ValueCompare) []models.ValueCompare {
	result := make([]models.ValueCompare, 0)
	for _, originalCompare := range compares {
		compare := originalCompare
		compare.Different = compare.RightOriginal != compare.LeftOriginal
		result = append(result, compare)
	}
	return result
}
