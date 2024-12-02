package test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"net/url"
	"path"
	"testing"
	"url_shortener/httpServer/handlers/url/random"
	"url_shortener/httpServer/handlers/url/save"
	"url_shortener/internal/lib/api"

	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())
	token := simulateLoginAndGetToken(t, e)

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.RandomString(10),
		}).
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

//nolint:funlen
func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())
			// login and get token
			token := simulateLoginAndGetToken(t, e)

			// Save
			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithHeader("Authorization", "Bearer "+token).
				Expect().
				Status(200).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")

				resp.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()

				alias = resp.Value("alias").String().Raw()
			}

			// Redirect
			testRedirect(t, alias, tc.url)

			// Delete
			reqDel := e.DELETE("/"+path.Join("url", alias)).
				WithHeader("Authorization", "Bearer "+token).
				Expect().
				Status(200).
				JSON().Object()

			reqDel.Value("status").String().IsEqual("OK")

			if tc.alias == "" {
				return
			}

		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}
func simulateLoginAndGetToken(t *testing.T, e *httpexpect.Expect) string {
	// Simulate login with valid credentials
	resp := e.POST("/login").
		WithJSON(map[string]string{
			"username": "brazza",
			"password": "mypass",
		}).
		Expect().
		Status(200).
		JSON().Object()

	// Extract the token from the response
	token := resp.Value("token").String().Raw()
	require.NotEmpty(t, token, "JWT token should not be empty")
	return token
}
