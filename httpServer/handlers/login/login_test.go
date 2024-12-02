package login_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url_shortener/httpServer/handlers/login"
	mocks "url_shortener/httpServer/handlers/login/url_shortener/test"
	"url_shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestHandleLogin(t *testing.T) {
	mockHandler := mocks.NewLoginHandler(t)
	mockLogger := slogdiscard.NewDiscardLogger()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBehavior   func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful login",
			requestBody: login.User{
				Username: "testuser",
				Password: "password123",
			},
			mockBehavior: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockHandler.On("GetUserByUsername", "testuser").Return(login.User{
					ID:       "12345",
					Username: "testuser",
					Password: string(hashedPassword),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"token":`, // Check for the presence of the token key
		},
		{
			name:           "Invalid JSON payload",
			requestBody:    `{"username": "testuser",`, // Malformed JSON
			mockBehavior:   nil,                        // No mock behavior needed
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"could not decode"`,
		},
		{
			name: "User not found",
			requestBody: login.User{
				Username: "nonexistent",
				Password: "password123",
			},
			mockBehavior: func() {
				mockHandler.On("GetUserByUsername", "nonexistent").Return(login.User{}, sql.ErrNoRows).Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"invalid username or password"`,
		},
		{
			name: "Invalid password",
			requestBody: login.User{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockBehavior: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockHandler.On("GetUserByUsername", "testuser").Return(login.User{
					ID:       "12345",
					Username: "testuser",
					Password: string(hashedPassword),
				}, nil).Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"invalid username or password"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			if tt.mockBehavior != nil {
				tt.mockBehavior()
			}

			var req *http.Request
			switch body := tt.requestBody.(type) {
			case login.User:
				jsonBody, err := json.Marshal(body)
				require.NoError(t, err)
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			case string:
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(body)))
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := login.HandleLogin(mockLogger, mockHandler)

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
