package server

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/gork74/aws-parameter-bulk/pkg/util"
	"github.com/rs/zerolog/log"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gork74/aws-parameter-bulk/conf"
	"github.com/rs/zerolog"
)

type MockSSM struct {
	ssmiface.SSMAPI
	err error
}

func nameString(parameter ssm.GetParametersInput) string {
	result := "["
	for i, param := range parameter.Names {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s", *param)
	}
	return result + "]"
}

func (sp *MockSSM) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	output := new(ssm.GetParametersOutput)
	log.Info().Msgf("%s", nameString(*input))
	if nameString(*input) == "[One1]" {
		name1 := "One1"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal1")})
	}
	if nameString(*input) == "[Three1]" {
		name1 := "Three1"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("ThreeVal1")})
	}
	if nameString(*input) == "[One2]" {
		name1 := "One2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal2")})
	}
	if nameString(*input) == "[One1, One2]" {
		name1 := "One1"
		name2 := "One2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal1")})
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name2, Value: aws.String("OneVal2")})
	}
	if nameString(*input) == "[Three1, Three2]" {
		name1 := "Three1"
		name2 := "Three2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("ThreeVal1")})
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name2, Value: aws.String("ThreeVal2")})
	}
	return output, sp.err
}

func (sp *MockSSM) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	output := new(ssm.GetParametersByPathOutput)
	params := make([]*ssm.Parameter, 0)
	if *input.Path == "/path" {
		params = append(params, &ssm.Parameter{Name: aws.String("One1"), Value: aws.String("OneVal1")})
	}
	if *input.Path == "/path2" {
		params = append(params, &ssm.Parameter{Name: aws.String("One1"), Value: aws.String("OneVal1")})
		params = append(params, &ssm.Parameter{Name: aws.String("One2"), Value: aws.String("OneVal2")})
	}
	output.Parameters = params
	return output, sp.err
}

// Define a custom testServer type which anonymously embeds a httptest.Server
// instance.
type testServer struct {
	*httptest.Server
}

// Define a regular expression which captures the CSRF token value from the
// HTML.
var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t *testing.T, body []byte) string {
	// Use the FindSubmatch method to extract the token from the HTML body.
	// Note that this returns an array with the entire matched pattern in the
	// first position, and the values of any captured data in the subsequent
	// positions.
	matches := csrfTokenRX.FindSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}

// Create a newTestApplication helper which returns an instance of our
// application struct containing mocked dependencies.
func newTestApplication(t *testing.T) *application {

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Caller().Logger()

	// Create an instance of the template cache.
	templateCache, err := newTemplateCache("./../ui/html/")
	if err != nil {
		t.Fatal(err)
	}

	conf.BindEnv()

	InitTemplates()
	session := scs.New()
	session.Lifetime = 24 * 30 * time.Hour

	ssmClient := util.NewSSM()
	ssmClient.SSM = &MockSSM{
		err: nil, //errors.New("my custom error"),
	}
	// Initialize the dependencies, using the mocks for the loggers and
	// the client.
	app := &application{
		logger:        &logger,
		session:       session,
		templateCache: templateCache,
		ssmClient:     ssmClient,
	}

	return app
}

// Create a newTestServer helper which initalizes and returns a new instance
// of our custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	// Initialize a new cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the client, so that response cookies are stored
	// and then sent with subsequent requests.
	ts.Client().Jar = jar

	// Disable redirect-following for the client. Essentially this function
	// is called after a 3xx response is received by the client, and returning
	// the http.ErrUseLastResponse error forces it to immediately return the
	// received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// Implement a get method on our custom testServer type. This makes a GET
// request to a given url path on the test server, and returns the response
// status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string, wantBody bool) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	if wantBody {
		body, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			t.Fatal(err)
		}

		return rs.StatusCode, rs.Header, body
	}

	return rs.StatusCode, rs.Header, nil
}

// Create a postForm method for sending POST requests to the test server.
// The final parameter to this method is a url.Values object which can contain
// any data that you want to send in the request body.
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values, wantBody bool) (int, http.Header, []byte) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body.
	defer rs.Body.Close()
	if wantBody {
		body, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			t.Fatal(err)
		}
		// Return the response status, headers and body.
		return rs.StatusCode, rs.Header, body
	}

	// Return the response status, headers.
	return rs.StatusCode, rs.Header, nil
}
