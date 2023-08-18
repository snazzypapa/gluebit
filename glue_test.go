package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetPortFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "portfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(dir, "portfile")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteString(`{"port": 12345}`)
	if err != nil {
		t.Error(err)
	}

	tt := []struct {
		name     string
		path     string
		portFile []byte
		expected int
		hasErr   bool
	}{
		{
			name:     "valid",
			path:     file.Name(),
			portFile: []byte(`{"port": 12345}`),
			expected: 12345,
			hasErr:   false,
		},
		{
			name:     "File does not exist",
			path:     "DNE",
			portFile: nil,
			expected: 0,
			hasErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			port, err := getPortFile(tc.path)
			if tc.hasErr && err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Fatal(err)
			}
			if port != tc.expected {
				t.Errorf("Expected port %d, got %d", tc.expected, port)
			}
		})
	}
}

func TestDecodeGlueTunPort(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		portFile []byte
		expected int
		hasErr   bool
	}{
		{
			name:     "valid",
			portFile: []byte(`{"port": 12345}`),
			expected: 12345,
			hasErr:   false,
		},
		{
			name:     "invalid",
			portFile: []byte(`Bingo Bango Bongo`),
			expected: 0,
			hasErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			port, err := decodeGlueTunPort(bytes.NewReader(tc.portFile))
			if tc.hasErr && err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Fatal(err)
			}
			if port != tc.expected {
				t.Errorf("Expected port %d, got %d", tc.expected, port)
			}
		})
	}
}

func TestGetPortApi(t *testing.T) {
	t.Parallel()

	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/openvpn/portforwarded" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"port": %d}`, 12345)))
		} else {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	tt := []struct {
		name     string
		url      string
		portFile []byte
		expected int
		hasErr   bool
	}{
		{
			name:     "valid",
			url:      server.URL,
			portFile: []byte(`{"port": 12345}`),
			expected: 12345,
			hasErr:   false,
		},
		{
			name:     "invalid protocol",
			url:      ":12345",
			expected: 0,
			hasErr:   true,
		},
		{
			name:     "invalid url",
			url:      "http://:12345",
			expected: 0,
			hasErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			port, err := getPortApi(tc.url, http.DefaultClient)
			if tc.hasErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("Unexpected error, %s", err)
			}
			if port != tc.expected {
				t.Errorf("Expected port %d, got %d", tc.expected, port)
			}
		})
	}
}

type mockDoer struct {
	err  error
	body string
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}
	return resp, m.err
}

func TestGetPort(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp("", "portfile")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(`{"port": 12345}`)
	if err != nil {
		t.Error(err)
	}
	tt := []struct {
		name     string
		httpDoer mockDoer
		config   Config
		wantPort int
		wantErr  bool
	}{
		{
			name: "valid",
			httpDoer: mockDoer{
				err:  nil,
				body: `{"port": 12345}`,
			},
			config: Config{
				GlueTunHost: "http://localhost",
				GlueTunPort: 8000,
			},
			wantPort: 12345,
			wantErr:  false,
		},
		{
			name: "Api error, no port file",
			httpDoer: mockDoer{
				err:  errors.New("test error"),
				body: `{"port": 12345}`,
			},
			config: Config{
				GlueTunHost:     "http://localhost",
				GlueTunPort:     8000,
				GlueTunPortFile: "",
			},
			wantPort: 0,
			wantErr:  true,
		},
		{
			name: "No api, file error",
			config: Config{
				GlueTunHost:     "",
				GlueTunPort:     0,
				GlueTunPortFile: "DNE",
			},
			wantPort: 0,
			wantErr:  true,
		},
		{
			name: "No api, file Ok",
			config: Config{
				GlueTunHost:     "",
				GlueTunPort:     0,
				GlueTunPortFile: file.Name(),
			},
			wantPort: 12345,
			wantErr:  false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := glueGetter{}
			port, err := g.GetGlueTunPort(tc.config, &tc.httpDoer)
			if tc.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error, %s", err)
			}
			if port != tc.wantPort {
				t.Errorf("Expected port %d, got %d", tc.wantPort, port)
			}
		})
	}
}
