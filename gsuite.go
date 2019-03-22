package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/reports/v1"
)

var googleAPIScopes = []string{
	"https://www.googleapis.com/auth/admin.reports.audit.readonly",
}

func setupGoogleClient(configData, tokenData []byte) (*http.Client, error) {
	config, err := google.ConfigFromJSON(configData, googleAPIScopes...)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse client secret file to config")
	}

	tok := oauth2.Token{}
	if err = json.Unmarshal(tokenData, &tok); err != nil {
		return nil, errors.Wrap(err, "Fail to parse oauth token data")
	}

	// saveToken("token.json", &tok)
	client := config.Client(context.Background(), &tok)
	return client, nil
}

type queue struct {
	err       error
	data      []byte
	timestamp time.Time
	key       string
	app       string
}

type application struct {
	name  string
	delta time.Duration
}

func exportAppLogs(ch chan *queue, srv *admin.Service, app *application, now time.Time) {
	var pageToken string
	timeFmt := "2006-01-02T15:04:05.000Z"

	for {
		query := srv.Activities.List("all", app.name).
			StartTime(now.Add(-app.delta).Format(timeFmt)).
			EndTime(now.Format(timeFmt)).
			MaxResults(1000)

		if pageToken != "" {
			query = query.PageToken(pageToken)
		}

		r, err := query.Do()
		if err != nil {
			ch <- &queue{err: errors.Wrap(err, "Unable to retrieve logins to domain.")}
			return
		}

		for _, item := range r.Items {
			q := new(queue)
			if q.data, err = item.MarshalJSON(); err != nil {
				ch <- &queue{err: errors.Wrap(err, "Fail to marshal AdminActivity.")}
				return
			}

			if q.timestamp, err = time.Parse(timeFmt, item.Id.Time); err != nil {
				ch <- &queue{err: errors.Wrapf(err, "Fail to parse timestamp: %s", item.Id.Time)}
				return
			}

			rawID, err := item.Id.MarshalJSON()
			if err != nil {
				ch <- &queue{err: errors.Wrap(err, "Fail to marshal ID of item")}
				return
			}

			q.key = fmt.Sprintf("%x", sha256.Sum256(rawID))
			q.app = item.Id.ApplicationName

			ch <- q
		}

		pageToken = r.NextPageToken

		if pageToken == "" {
			break
		}
	}
}

func exportLogs(client *http.Client) chan *queue {
	ch := make(chan *queue)

	apps := []application{
		{"admin", time.Minute * 10},
		{"drive", time.Minute * 10},
		{"mobile", time.Minute * 10},
		{"token", time.Minute * 10},
		{"login", time.Hour * 48},
	}

	now := time.Now().UTC()

	go func() {
		srv, err := admin.New(client)
		if err != nil {
			ch <- &queue{err: errors.Wrap(err, "Unable to retrieve reports Client")}
			return
		}

		wg := &sync.WaitGroup{}

		for idx := range apps {
			wg.Add(1)

			go func(app *application) {
				exportAppLogs(ch, srv, app, now)
				wg.Done()
			}(&apps[idx])
		}

		wg.Wait()
		close(ch)
	}()

	return ch
}
