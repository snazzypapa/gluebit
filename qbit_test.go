package main

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/publicsuffix"
)

func TestClient_Login(t *testing.T) {
	t.Parallel()

	defaultTimeout = time.Duration(60 * time.Second)
	correctUser := "testuser"
	correctPass := "testpass"
	// Create a test server to mock the qBittorrent API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			// Check that the request body contains the correct username and password
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			// Fail if the username or password is empty
			if r.FormValue("username") == "" || r.FormValue("password") == "" {
				t.Fatalf("unexpected form values: %v", r.Form)
			}

			// Return a successful response with a cookie
			if r.FormValue("username") == correctUser && r.FormValue("password") == correctPass {
				http.SetCookie(w, &http.Cookie{Name: "testcookie", Value: "testvalue"})
				w.Write([]byte(ResponseBodyOK))
				return
			}

			// Return a failed response
			w.Write([]byte(ResponseBodyFAIL))
		} else {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	tt := []struct {
		name     string
		username string
		password string
		hasErr   bool
	}{
		{
			name:     "valid",
			username: correctUser,
			password: correctPass,
			hasErr:   false,
		},
		{
			name:     "invalid username",
			username: "baduser",
			password: correctPass,
			hasErr:   true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new client with the test server URL
			client, err := NewClient(ts.URL, tc.username, tc.password)
			if tc.hasErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("Unexpected error, %s", err)
			}

			if tc.hasErr {
				return
			}

			// Check that the cookie was added to the client's cookie jar
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}
			u.Path = ""
			cookies := client.Jar.Cookies(u)
			if len(cookies) != 1 || cookies[0].Name != "testcookie" || cookies[0].Value != "testvalue" {
				t.Fatalf("unexpected cookies: %v", cookies)
			}
		})
	}
}

func TestClient_GetPreferences(t *testing.T) {
	t.Parallel()

	defaultTimeout = time.Duration(60 * time.Second)
	// Create a test server to mock the qBittorrent API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/app/preferences" {
			// Return a successful response with a preferences object
			prefs := Preferences{ListenPort: 1234}
			b, err := json.Marshal(prefs)
			if err != nil {
				t.Fatalf("failed to marshal preferences: %v", err)
			}
			w.Write(b)
		} else {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	cliJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &Client{
		&http.Client{
			Jar: cliJar,
		},
		ts.URL + "/api/v2/",
	}

	// Call the GetPreferences method on the client
	prefs, err := client.GetPreferences()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the preferences object was returned correctly
	if prefs.ListenPort != 1234 {
		t.Fatalf("unexpected preferences: %v", prefs)
	}
}

func TestClient_SetPreferences(t *testing.T) {
	t.Parallel()

	defaultTimeout = time.Duration(60 * time.Second)
	newPort := 9999
	// Create a test server to mock the qBittorrent API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/app/setPreferences" {
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			var prefs Preferences
			form := r.FormValue("json")
			if err := json.NewDecoder(strings.NewReader(form)).Decode(&prefs); err != nil {
				t.Fatalf("failed to decode preferences: %v", err)
			}

			// check listen port
			if prefs.ListenPort != newPort {
				t.Fatalf("unexpected preferences: %v", prefs)
			}
			// Return a successful response
			w.Write([]byte(ResponseBodyOK))
		} else {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	cliJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &Client{
		&http.Client{
			Jar: cliJar,
		},
		ts.URL + "/api/v2/",
	}

	// Call the GetPreferences method on the client
	err := client.SetPreferences(Preferences{ListenPort: newPort})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
