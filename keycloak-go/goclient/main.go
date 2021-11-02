package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	clientID     = "myclient"
	clientSecret = "2ca0707c-f127-4fdf-abe3-d58eef2c8222"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/auth/realms/myrealm")

	if err != nil {
		log.Fatalf(err.Error())
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8081/auth/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	state := "123"

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(rw, "Invalid state", http.StatusBadRequest)
			return
		}

		//Token de Autorização
		code := r.URL.Query().Get("code")
		token, err := config.Exchange(ctx, code)
		if err != nil {
			http.Error(rw, "Failed to change token", http.StatusInternalServerError)
			return
		}

		//Token de Autenticação
		idToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(rw, "Failed to generate id_token", http.StatusInternalServerError)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err != nil {
			http.Error(rw, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		resp := struct {
			AccessToken *oauth2.Token
			IDToken     string
			UserInfo    *oidc.UserInfo
		}{
			token,
			idToken,
			userInfo,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(data)

	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
