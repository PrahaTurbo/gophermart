package storage

import "errors"

const uniqueViolationErrCode = "23505"

var ErrAlreadyExist = errors.New("login already exist in database")
