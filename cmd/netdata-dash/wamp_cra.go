package main

import "errors"

type serverKeyStore struct {
	provider string
	role     string

	pbkdf2Key  []byte
	keylen     int
	iterations int

	salt   string
	ticket []byte
}

func (ks *serverKeyStore) AuthKey(authid, authmethod string) ([]byte, error) {
	if authid != "jdoe" {
		return nil, errors.New("no such user: " + authid)
	}
	switch authmethod {
	case "wampcra":
		// Lookup the user's PBKDF2-derived key.
		return ks.pbkdf2Key, nil
	case "ticket":
		// Lookup the user's key.
		return ks.ticket, nil
	}
	return nil, errors.New("unsupported authmethod")
}

func (ks *serverKeyStore) PasswordInfo(authid string) (string, int, int) {
	if authid != "jdoe" {
		return "", 0, 0
	}
	return ks.salt, ks.keylen, ks.iterations
}

func (ks *serverKeyStore) Provider() string { return ks.provider }

func (ks *serverKeyStore) AuthRole(authid string) (string, error) {
	if authid != "jdoe" {
		return "", errors.New("no such user: " + authid)
	}
	return ks.role, nil
}
