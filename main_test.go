package main

import (
	"errors"
	"net/http"
	"testing"
)

type mockClient struct {
	pref        Preferences
	getPrefsErr error
	setPrefsErr error
}

func (m *mockClient) GetPreferences() (Preferences, error) {
	return m.pref, m.getPrefsErr
}

func (m *mockClient) SetPreferences(p Preferences) error {
	m.pref = p
	return m.setPrefsErr
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return nil, m.getPrefsErr
}

type mockGlueGetter struct {
	port int
	err  error
	runs int
}

func (m *mockGlueGetter) GetGlueTunPort(Config, HttpDoer) (int, error) {
	m.runs++
	if m.err != nil {
		return 0, m.err
	}
	return m.port, m.err
}

func TestSetPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		client     *mockClient
		glue       *mockGlueGetter
		wantErr    bool
		wantPort   int
		wantRandom bool
	}{
		{
			name: "port already set",
			client: &mockClient{
				pref: Preferences{
					ListenPort: 1234,
				},
			},
			glue: &mockGlueGetter{
				port: 1234,
			},
			wantPort: 1234,
			wantErr:  false,
		},
		{
			name: "different port set",
			client: &mockClient{
				pref: Preferences{
					ListenPort: 9999,
				},
			},
			glue: &mockGlueGetter{
				port: 1234,
			},
			wantPort: 1234,
			wantErr:  false,
		},
		{
			name: "get gluetun port error",
			client: &mockClient{
				pref: Preferences{
					ListenPort: 1234,
				},
			},
			glue: &mockGlueGetter{
				err: errors.New("glue error"),
			},
			wantErr: true,
		},
		{
			name: "get prefs error",
			client: &mockClient{
				getPrefsErr: errors.New("client error"),
			},
			glue: &mockGlueGetter{
				port: 1234,
			},
			wantErr: true,
		},
		{
			name: "get prefs error",
			client: &mockClient{
				pref: Preferences{
					ListenPort: 9999,
				},
				setPrefsErr: errors.New("client error"),
			},
			glue: &mockGlueGetter{
				port: 1234,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setPort(Config{}, tt.client, tt.glue)
			if (err != nil) != tt.wantErr {
				t.Errorf("setPort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantPort != 0 && tt.client.pref.ListenPort != tt.wantPort {
				t.Errorf("setPort() port = %v, wantPort %v", tt.client.pref.ListenPort, tt.wantPort)
			}
		})
	}
}

// func TestRun(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name     string
// 		client   *mockClient
// 		glue     *mockGlueGetter
// 		wantErr  bool
// 		config   Config
// 		wantRuns int
// 	}{
// 		{
// 			name: "port already set",
// 			client: &mockClient{
// 				pref: Preferences{
// 					ListenPort: 1234,
// 				},
// 			},
// 			glue: &mockGlueGetter{
// 				port: 1234,
// 			},
// 			wantErr: false,
// 			config: Config{
// 				UpdateInterval: 0,
// 			},
// 			wantRuns: 1,
// 		},
// 		{
// 			name: "return error",
// 			glue: &mockGlueGetter{
// 				err: errors.New("glue error"),
// 			},
// 			wantErr: true,
// 			config: Config{
// 				UpdateInterval: 0,
// 			},
// 			wantRuns: 1,
// 		},
// 		{
// 			name: "run continuously",
// 			client: &mockClient{
// 				pref: Preferences{
// 					ListenPort: 1234,
// 				},
// 			},
// 			glue: &mockGlueGetter{
// 				port: 1234,
// 			},
// 			wantErr: false,
// 			config: Config{
// 				UpdateInterval: 1,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(
// 				context.Background(),
// 				time.Duration(2300*time.Millisecond), // just over two seconds
// 			)
// 			defer cancel()
// 			err := run(ctx, tt.config, tt.client, tt.glue)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("setPort() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if tt.wantRuns != 0 && tt.glue.runs != tt.wantRuns {
// 				t.Errorf("Want runs setPort() runs = %v, wantRuns %v", tt.glue.runs, tt.wantRuns)
// 			}
// 			if tt.config.UpdateInterval > 0 && tt.glue.runs < 2 {
// 				t.Errorf("setPort() runs = %v, wantRuns %v", tt.glue.runs, tt.wantRuns)
// 			}
// 		})
// 	}
// }
