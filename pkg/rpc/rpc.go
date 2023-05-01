package rpc

import (
	"io"
	"net/http"
)

var client = &http.Client{}

func Delete(remote string) error {
	request, _ := http.NewRequest(http.MethodDelete, remote, nil)

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return &RequestError{
			Method:     http.MethodDelete,
			StatusCode: response.StatusCode,
		}
	}

	return nil
}

func Get(remote string) (string, error) {
	request, _ := http.NewRequest(http.MethodGet, remote, nil)

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", &RequestError{
			Method:     http.MethodGet,
			StatusCode: response.StatusCode,
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func Head(remote string) error {
	request, _ := http.NewRequest(http.MethodHead, remote, nil)

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return &RequestError{
			Method:     http.MethodHead,
			StatusCode: response.StatusCode,
		}
	}

	return nil
}

func Put(remote string, length int64, body io.Reader) error {
	request, _ := http.NewRequest(http.MethodPut, remote, body)
	request.ContentLength = length

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusNoContent {
		return &RequestError{
			Method:     http.MethodPut,
			StatusCode: response.StatusCode,
		}
	}

	return nil
}
