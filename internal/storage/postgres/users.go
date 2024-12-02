package postgres

import (
	"context"

	"url_shortener/httpServer/handlers/login"
)

func (s *Storage) GetUserByUsername(username string) (login.User, error) {
	var user login.User
	row := s.DB.QueryRow(context.Background(), `SELECT id, username, password FROM users WHERE username = $1`, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return login.User{}, err
	}

	return user, nil
}
