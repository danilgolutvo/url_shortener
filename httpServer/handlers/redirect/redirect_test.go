package redirect

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	mocks2 "url_shortener/httpServer/handlers/redirect/mocks"
	"url_shortener/internal/lib/api"
	"url_shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "valid alias",
			alias: "google",
			url:   "https://google.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlGetterMock := mocks2.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			router := mux.NewRouter()
			router.Handle("/{alias}", New(slogdiscard.NewDiscardLogger(), urlGetterMock)).Methods(http.MethodGet)

			ts := httptest.NewServer(router)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			// Check the final URL after redirection.
			assert.Equal(t, tc.url, redirectedToURL)
		})
	}
}
