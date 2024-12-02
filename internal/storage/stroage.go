package storage

import "errors"

var ErrURLNotFound = errors.New("url not found")
var ErrURLExists = errors.New("url exists")
var ErrCaseMismatch = errors.New("case mismatch")
var ErrAliasNotFound = errors.New("alias not found")
