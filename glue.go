package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"
)

type glueGetter struct {
}

func (g glueGetter) GetGlueTunPort(config Config, requester HttpDoer) (int, error) {
	return getPort(config, requester)
}

func decodeGlueTunPort(toRead io.Reader) (int, error) {
	var portFile port
	decoder := json.NewDecoder(toRead)
	err := decoder.Decode(&portFile)
	return portFile.Port, err
}

// getPortFile returns the forwarded port from a file written by gluetun.
func getPortFile(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return decodeGlueTunPort(file)
}

// getPortApi returns the forwarded port from gluetun's api.
func getPortApi(url string, client HttpDoer) (int, error) {
	url = url + "/v1/openvpn/portforwarded"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return decodeGlueTunPort(resp.Body)
}

// getPort returns the forwarded port from gluetun.
func getPort(config Config, client HttpDoer) (int, error) {
	var port int
	var apiErr error
	var fileErr error

	if config.GlueTunPort != 0 {
		port, apiErr = getPortApi(config.gluetunUrl(), client)
		if apiErr == nil {
			return port, nil
		}
	}
	if config.GlueTunPortFile == "" {
		return 0, apiErr
	}
	port, fileErr = getPortFile(config.GlueTunPortFile)
	if fileErr == nil {
		return port, nil
	}

	return 0, errors.Join(apiErr, fileErr)
}
