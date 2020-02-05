package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Login into sporteasy and return the cookie with the token
// The cookie will be used for later requests
func login(u *url.URL, username, password, csrfToken string, cookies []*http.Cookie) (int, *http.Cookie) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	jar.SetCookies(u, cookies)
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{Jar: jar}
	if *debug {
		client = http.Client{
			Jar:       jar,
			Transport: &logTransport{http.DefaultTransport},
		}
	}

	values := make(url.Values)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("csrfmiddlewaretoken", csrfToken)
	values.Set("next", "https://www.sporteasy.net/fr/profile/")
	values.Set("invitation_token", "")
	values.Set("detectRedirect", "true")

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(values.Encode()))

	csrfCookie := ""
	for _, c := range cookies {
		if c.Name == "se_csrftoken" {
			csrfCookie = c.Value
		}
	}
	req.Header.Set("X-CSRFToken", csrfCookie)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode < 400 {
		return resp.StatusCode, nil
	}

	c := jar.Cookies(u)
	return resp.StatusCode, c[0]
}

func followUrlWithCookies(u *url.URL) ([]byte, []*http.Cookie, error) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{Jar: jar}
	if *debug {
		client = http.Client{
			Jar:       jar,
			Transport: &logTransport{http.DefaultTransport},
		}
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return []byte{}, []*http.Cookie{}, err
	}

	resp, err := client.Do(req)
	if resp.StatusCode > 400 {
		return []byte{}, []*http.Cookie{}, fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, []*http.Cookie{}, err
	}

	return body, jar.Cookies(u), nil
}

func followUrl(okUrl string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", okUrl, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := client.Do(req)
	if resp.StatusCode > 400 {
		return []byte{}, fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}
