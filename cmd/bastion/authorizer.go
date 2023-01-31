package main

import "github.com/toddgaunt/bastion"

type simpleAuthorizer struct {
	username string
	password string
}

func (sa simpleAuthorizer) Authorize(username, password string) (bastion.Claims, bool) {
	// Don't allow either empty usernames or passwords
	if sa.username == "" || sa.password == "" {
		return bastion.Claims{}, false
	}

	if sa.username == username && sa.password == password {
		return bastion.Claims{
			Username: username,
			Admin:    true,
		}, true
	}

	return bastion.Claims{}, false
}
