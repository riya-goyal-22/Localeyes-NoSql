package utils

import "errors"

var NotYourPost = errors.New("no post of yours exist with this id")
var NotYourQuestion = errors.New("no question of yours exist with this id")
var NoPost = errors.New("no post exist with this id")
var NoQuestion = errors.New("no question exist with this id")
var NoUser = errors.New("no user exist")
var TitleMissing = errors.New("required field 'title' is missing")
var ContentMissing = errors.New("required field 'content' is missing")
var TypeMissing = errors.New("required field 'type' is missing")
var InvalidPost = errors.New("invalid post type")
var InvalidAnswer = errors.New("invalid answer")
var InvalidAccountCredentials = errors.New("invalid account credentials")
var InactiveUser = errors.New("inactive user")
var WrongOTP = errors.New("wrong otp")
var UserExistsEmail = errors.New("user exists with this email")
var UserExistsName = errors.New("user exists with this username")
