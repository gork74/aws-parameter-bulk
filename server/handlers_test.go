package server

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

func Test_application_getHome(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		wantCode int
		wantBody []byte
	}{
		{"Get home", http.StatusOK, []byte("Load and Compare")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			code, _, body := ts.get(t, "/", tt.wantBody != nil)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if tt.wantBody != nil && !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}

func Test_application_postHome(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Make a GET /request and then extract the CSRF token from the
	// response body.
	_, _, body := ts.get(t, "/", true)
	csrfToken := extractCSRFToken(t, body)

	t.Log(csrfToken)

	tests := []struct {
		name          string
		namesleft     string
		nameright     string
		jsonleft      bool
		recursiveleft bool
		csrfToken     string
		wantCode      int
		wantBody      []byte
	}{
		{"Empty namesleft and namesright", "", "", true, true, csrfToken, http.StatusOK, []byte("Names not valid")},
		{"Valid namesleft", "Three1", "", false, true, csrfToken, http.StatusOK, []byte(">ThreeVal1<")},
		{"Empty namesleft but namesright", "", "One1", false, true, csrfToken, http.StatusOK, []byte("Names not valid")},
		{"Valid namesleft and namesright", "One1", "One2", false, true, csrfToken, http.StatusOK, []byte(">OneVal2<")},
		{"Invalid CSRF Token", "One1", "", false, true, "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("namesleft", tt.namesleft)
			form.Add("namesright", tt.nameright)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/", form, tt.wantBody != nil)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if tt.wantBody != nil && !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}

func Test_application_postReset(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Make a GET /request and then extract the CSRF token from the
	// response body.
	_, _, body := ts.get(t, "/", true)
	csrfToken := extractCSRFToken(t, body)

	t.Log(csrfToken)

	tests := []struct {
		name      string
		csrfToken string
		wantCode  int
		wantBody  []byte
	}{
		{"Empty textarea", csrfToken, http.StatusOK, []byte("></textarea>")},
		{"Invalid CSRF Token", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/reset", form, tt.wantBody != nil)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if tt.wantBody != nil && !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}
