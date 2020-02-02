package main

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetUrl(t *testing.T) {
	content, err := ioutil.ReadFile("test.text")
	if err != nil {
		t.Fatal(err)
	}

	url, _ := extractValidationUrl(strings.NewReader(string(content)))
	if len(url) == 0 {
		t.Fatal(errors.New("empty node"))
	}

}

func TestGetRedirectUrl(t *testing.T) {
	content, err := ioutil.ReadFile("test2.txt")
	if err != nil {
		t.Fatal(err)
	}

	url, _ := extractRedirectUrl(strings.NewReader(string(content)))
	if len(url) == 0 {
		t.Fatal(errors.New("no redirect nde"))
	}
}

func TestGetPlayers(t *testing.T) {
	content, err := ioutil.ReadFile("test3.txt")
	if err != nil {
		t.Fatal(err)
	}

	url, _ := parseAttendeePage(strings.NewReader(string(content)))
	if len(url) == 0 {
		t.Fatal(errors.New("no redirect nde"))
	}
}
