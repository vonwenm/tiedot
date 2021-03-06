package main

import (
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"testing"
	"time"
)

var (
	privateKey []byte //openssl genrsa -out rsa.key 1024
	publicKey  []byte //openssl rsa -in rsa.key -out rsa.pub
	//openssl req -new -key rsa.key -out rsa.csr
	//openssl x509 -req -in rsa.csr -signkey rsa.key -out rsa.crt -days 365
)

func TestRsa(t *testing.T) {
	var err error
	if privateKey, err = ioutil.ReadFile("rsa.key"); err != nil {
		t.Fatal(err)
	}
	if publicKey, err = ioutil.ReadFile("rsa.pub"); err != nil {
		t.Fatal(err)
	}
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["PERMISSION"] = "admin@tiedot"
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	if ts, err := token.SignedString(privateKey); err != nil {
		t.Fatal(err)
	} else {
		check(t, ts)
	}
}

func check(t *testing.T, ts string) {
	var err error
	var token *jwt.Token
	if token, err = jwt.Parse(ts, func(ts *jwt.Token) (interface{}, error) {
		return publicKey, nil
	}); err != nil {
		t.Fatal(err)
	}
	if token.Valid {
		t.Log(token)
	} else {
		t.Log(token)
		t.Fail()
	}
}
