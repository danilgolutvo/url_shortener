package redirect

import (
	mux2 "github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"url_shortener/httpServer/handlers/redirect/mocks"
	"url_shortener/internal/lib/api"
	"url_shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://www.google.com/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			router := mux2.NewRouter()

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
