package main

import (
	"errors"
	"io"
	"log"
	"regexp"

	"golang.org/x/net/html"
)

const (
	firstPattern  = "Oui"
	secondPattern = "Click here if you"
)

// Extract the validation link from email body
func extractValidationUrl(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var url string
	var okNode html.Node
	findNode(doc, &okNode, regexp.MustCompile(firstPattern))

	//this node has a span parent which has, as parent, the node we are looking for
	if okNode.Parent == nil {
		return "", errors.New("url node not found")
	}

	urlNode := okNode.Parent.Parent
	for _, attr := range urlNode.Attr {
		if attr.Key == "href" {
			url = attr.Val
		}
	}

	if len(url) > 0 {
		return url, nil
	}

	return "", errors.New("link not found")

}

// extracts the url from the redirect page
func extractRedirectUrl(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var url string
	var redirectNode html.Node
	findNode(doc, &redirectNode, regexp.MustCompile("noscript"))

	//this node has a span parent which has, as parent, the node we are looking for
	if redirectNode.Parent == nil {
		return "", errors.New("url node not found")
	}

	data := redirectNode.FirstChild.Data
	found := regexp.MustCompile(`href="(?P<url>^$)"`).FindStringSubmatch(data)

	if len(found) > 0 {
		return url, nil
	}

	return "", errors.New("link not found")

}

func findNode(n *html.Node, found *html.Node, r *regexp.Regexp) {
	if r.MatchString(n.Data) {
		log.Print("found node")
		*found = *n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findNode(c, found, r)
	}
}
