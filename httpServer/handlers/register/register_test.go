package register

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url_shortener/httpServer/handlers/login"
	"url_shortener/httpServer/handlers/register/mocks"
	"url_shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleRegistration(t *testing.T) {
	mockHandler := mocks.NewRegistrationHandler(t)
	mockLogger := slogdiscard.NewDiscardLogger()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBehavior   func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful registration",
			requestBody: login.User{
				Username: "testuser",
				Password: "securepassword",
			},
			mockBehavior: func() {
				mockHandler.On("CreateUser", mock.MatchedBy(func(user login.User) bool {
					return user.Username == "testuser" && user.Password == "securepassword"
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"registration is done successfully JSON"`,
		},
		{
			name:        "Invalid JSON payload",
			requestBody: `{"username": "testuser",`, // Malformed JSON
			mockBehavior: func() {
				// No mock behavior; CreateUser should not be called
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"could not decode"`,
		},
		{
			name: "Failed to create user",
			requestBody: login.User{
				Username: "testuser",
				Password: "securepassword",
			},
			mockBehavior: func() {
				mockHandler.On("CreateUser", mock.MatchedBy(func(user login.User) bool {
					return user.Username == "testuser" && user.Password == "securepassword"
				})).Return(errDatabaseError).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"server error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tt.mockBehavior()

			var req *http.Request
			switch body := tt.requestBody.(type) {
			case login.User:
				jsonBody, err := json.Marshal(body)
				require.NoError(t, err)
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			case string:
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(body)))
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := HandleRegistration(mockLogger, mockHandler)

			// Act
			handler.ServeHTTP(rr, req)

			// Assert
			res := rr.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			body, _ := io.ReadAll(res.Body)
			require.Contains(t, string(body), tt.expectedBody)
		})
	}
}

// Example of an error to use in the test
var errDatabaseError = fmt.Errorf("database error")
