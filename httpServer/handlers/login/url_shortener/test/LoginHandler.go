// Code generated by mockery v2.49.1. DO NOT EDIT.

package mocks

import (
	login "url_shortener/httpServer/handlers/login"

	mock "github.com/stretchr/testify/mock"
)

// LoginHandler is an autogenerated mock type for the LoginHandler type
type LoginHandler struct {
	mock.Mock
}

// GetUserByUsername provides a mock function with given fields: username
func (_m *LoginHandler) GetUserByUsername(username string) (login.User, error) {
	ret := _m.Called(username)

	if len(ret) == 0 {
		panic("no return value specified for GetUserByUsername")
	}

	var r0 login.User
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (login.User, error)); ok {
		return rf(username)
	}
	if rf, ok := ret.Get(0).(func(string) login.User); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Get(0).(login.User)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewLoginHandler creates a new instance of LoginHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLoginHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *LoginHandler {
	mock := &LoginHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
