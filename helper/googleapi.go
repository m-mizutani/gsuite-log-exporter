package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleAPIScopes = []string{
	"https://www.googleapis.com/auth/admin.reports.audit.readonly",
}

func saveOAuthToken(oauthPath, tokenPath string) error {
	configData, err := ioutil.ReadFile(oauthPath)
	if err != nil {
		return errors.Wrap(err, "Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(configData, googleAPIScopes...)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	log.Printf("Saving credential file: %s\n", tokenPath)
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)

	return nil
}
