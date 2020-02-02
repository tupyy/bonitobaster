package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
