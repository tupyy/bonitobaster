package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gmail "google.golang.org/api/gmail/v1"
)

const (
	AKKABonito = "AKKA_Bonito"
	subject    = "Dispo pour le match entre nous du"
)

func init() {
	registerApp("gmail", gmail.MailGoogleComScope, gmailMain)
}

type message struct {
	size    int64
	gmailID string
	date    time.Time
	body    string
	snippet string
}

//   go build -o bonitobuster *.go
//   bonitobuster -clientid="my-clientid" -secret="my-secret"
func gmailMain(client *http.Client, argv []string) {
	if len(argv) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: gmail")
		return
	}

	svc, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v", err)
	}

	var total int64
	msgs := []message{}
	pageToken := ""
	for {
		req := svc.Users.Messages.List("me").Q(fmt.Sprintf("from:%s", AKKABonito)).Q(fmt.Sprintf("subject:%s", subject))
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve messages: %v", err)
		}

		log.Printf("Processing %v messages...\n", len(r.Messages))
		for _, m := range r.Messages {
			msg, err := svc.Users.Messages.Get("me", m.Id).Format("full").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve message %v: %v", m.Id, err)
			}

			var date time.Time
			for _, h := range msg.Payload.Headers {
				if h.Name == "Date" {
					date, err = time.Parse(time.RFC1123Z, h.Value)
					if err != nil {
						log.Fatalf("Unable to parse date from header %v: %v", m.Id, err)
					}
					break
				}
			}

			// get body
			var html string
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/html" {
					data, _ := base64.URLEncoding.DecodeString(part.Body.Data)
					html = string(data)
				}
			}

			if isToday(date) {
				msgs = append(msgs, message{
					size:    msg.SizeEstimate,
					gmailID: msg.Id,
					date:    date,
					body:    html,
					snippet: msg.Snippet,
				})
			}
		}

		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
	}
	log.Printf("total: %v\n", total)

	for _, m := range msgs {
		fmt.Printf("Size: %v, Date: %v, Snippet: %q\n", m.size, m.date, m.snippet)
		log.Println("Extract validation url")
		url, err := extractValidationUrl(strings.NewReader(m.body))
		if err != nil {
			log.Fatalf("Url not found in message %v: %v", m.gmailID, m)
		}

		content, err := followUrl(url)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Extract redirect url")
		redirectUrl, err := extractRedirectUrl(strings.NewReader(string(content)))
		if err != nil {
			log.Fatalf("Redirect url not found in response from validation link: %s", string(content))
		}

		content2, err := followUrl(redirectUrl)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(content2))
	}
}

func isToday(d time.Time) bool {
	return true
	//	return d.Day() == time.Now().Day() && d.Month() == time.Now().Month() && d.Year() == time.Now().Year()
}
