package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	gmail "google.golang.org/api/gmail/v1"
)

const (
	AKKABonito = "AKKA_Bonito"
	subject    = "Dispo pour le match entre nous du"
	username   = "cosmin.tupangiu@gmail.com"
	password   = "Parola001."
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

func gmailMain(client *http.Client, argv []string) {
	if len(argv) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: gmail")
		return
	}

	svc, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v", err)
	}

	msgs, err := getMessages(svc)
	if err != nil {
		log.Fatal(err)
	}

	if len(msgs) > 0 {
		// login
		u, _ := url.Parse("https://www.sporteasy.net/fr/login/")
		content, cookies, err := followUrlWithCookies(u)
		if err != nil {
			log.Fatal(err)
		}

		csrfToken, err := extractCsrfToken(strings.NewReader(string(content)))
		if err != nil {
			log.Fatal(err)
		}

		statusCode, cookie := login(u, username, password, csrfToken, cookies)
		if statusCode > 400 {
			log.Fatalf("Cannot login. Status code: %d", statusCode)
		}
		log.Print(cookie.String())
		for _, m := range msgs {
			err := processMessage(m)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		}
	} else {
		log.Print("No messages yet..")
	}
}

// Retrive all the messages from AKKABonito received today
func getMessages(svc *gmail.Service) ([]message, error) {
	msgs := []message{}
	pageToken := ""
	for {
		req := svc.Users.Messages.List("me").Q(fmt.Sprintf("from:%s", AKKABonito)).Q(fmt.Sprintf("subject:%s", subject))
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			return []message{}, fmt.Errorf("Unable to retrieve messages: %v", err)
		}

		log.Printf("Processing %v messages...\n", len(r.Messages))
		for _, m := range r.Messages {
			msg, err := svc.Users.Messages.Get("me", m.Id).Format("full").Do()
			if err != nil {
				return []message{}, fmt.Errorf("Unable to retrieve message %v: %v", m.Id, err)
			}

			var date time.Time
			for _, h := range msg.Payload.Headers {
				if h.Name == "Date" {
					date, err = time.Parse(time.RFC1123Z, h.Value)
					if err != nil {
						return []message{}, fmt.Errorf("Unable to parse date from header %v: %v", m.Id, err)
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

	return msgs, nil
}

func isToday(d time.Time) bool {
	return d.Day() == time.Now().Day() && d.Month() == time.Now().Month() && d.Year() == time.Now().Year()
}
