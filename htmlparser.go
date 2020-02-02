package main

import (
	"errors"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const (
	okpattern = "Oui"
)

// Extract the validation link from email body
func extractValidationUrl(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	var url string
	var okNode html.Node
	findNode(doc, &okNode, regexp.MustCompile(okpattern))

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

	var redirectNode html.Node
	findNode(doc, &redirectNode, regexp.MustCompile("noscript"))

	//this node has a span parent which has, as parent, the node we are looking for
	if redirectNode.Parent == nil {
		return "", errors.New("url node not found")
	}

	data := redirectNode.FirstChild.Data
	found := regexp.MustCompile(`href="(?P<url>.+?)"`).FindStringSubmatch(data)

	if len(found) > 0 {
		return found[1], nil
	}

	return "", errors.New("link not found")

}

// parse the attendee page and return the list with attendee players
func parseAttendeePage(r io.Reader) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return []string{}, err
	}

	var div html.Node
	findNodeByAttribute(doc, &div, "data-attendance-group", "played")

	if div.Parent == nil {
		return []string{}, errors.New("category played div not found")
	}

	var ol html.Node
	findNodeByAttribute(&div, &ol, "class", "attendees")
	if ol.Parent == nil {
		return []string{}, errors.New("attendee div not found")
	}

	players := []string{}
	for c := ol.FirstChild; c != nil; c = c.NextSibling {
		var nameNode html.Node
		findNodeByAttribute(c, &nameNode, "class", "name")
		if nameNode.Parent != nil {
			players = append(players, cleanPlayerName(nameNode.FirstChild.Data))
		}
	}

	return players, nil
}

func findNode(n *html.Node, found *html.Node, r *regexp.Regexp) {
	if r.MatchString(n.Data) {
		*found = *n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findNode(c, found, r)
	}
}

func findNodeByAttribute(n *html.Node, found *html.Node, key, val string) {
	if n.Type == html.ElementNode {
		if len(n.Attr) > 0 {
			for _, a := range n.Attr {
				if a.Key == key && strings.TrimSpace(a.Val) == val {
					*found = *n
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findNodeByAttribute(c, found, key, val)
	}
}

func cleanPlayerName(name string) string {
	re := regexp.MustCompile(`\r?\n`)
	return strings.TrimSpace(re.ReplaceAllString(name, " "))
}
