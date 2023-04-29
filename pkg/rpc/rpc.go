package rpc

import (
	"io"
	"net/http"
)

func Put(remote string, length int64, body io.Reader) bool {
	req, _ := http.NewRequest(http.MethodPut, remote, body)
	req.ContentLength = length

	client := http.Client{}
	if resp, err := client.Do(req); err == nil {
		return resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent
	}

	return false
}

func Delete(remote string) bool {
	req, _ := http.NewRequest(http.MethodDelete, remote, nil)

	client := http.Client{}
	if resp, err := client.Do(req); err == nil {
		return resp.StatusCode == http.StatusNoContent
	}

	return false
}
