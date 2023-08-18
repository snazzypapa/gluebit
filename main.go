package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"
)

const loginAttempts = 10
const loginDelay = 10 * time.Second

type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Preferencer interface {
	GetPreferences() (Preferences, error)
	SetPreferences(Preferences) error
	HttpDoer
}

type GlueGetter interface {
	GetGlueTunPort(Config, HttpDoer) (int, error)
}

// setPort is the main function of the program.
// It gets the port from gluetun and sets it in qbittorrent.
func setPort(config Config, client Preferencer, glue GlueGetter) error {
	port, err := glue.GetGlueTunPort(config, client)
	if err != nil {
		return err
	}
	slog.Debug("Got port from gluetun", "port", port)
	pref, err := client.GetPreferences()
	if err != nil {
		return err
	}
	if pref.ListenPort == port {
		slog.Info("Port already set")
		return nil
	}
	pref.ListenPort = port
	pref.RandomPort = false
	err = client.SetPreferences(pref)
	if err != nil {
		return err
	}
	slog.Info("Set port to ", "port", port)
	return nil
}

// run runs the program in a loop.
func run(ctx context.Context, config Config, client Preferencer, glue GlueGetter) error {
	for {
		err := setPort(config, client, glue)
		if err != nil {
			slog.Warn("Failed to set port: ", err)
		}
		if config.UpdateInterval == 0 {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(config.UpdateInterval) * time.Second):
		}
	}
}

// getQbitClient returns a qbittorrent client.
func getQbitClient(config Config) *Client {
	tries := loginAttempts
	for {
		client, err := NewClient(config.qbitUrl(), config.QbitUsername, config.QbitPassword)
		if err == nil {
			return client
		}
		if errors.Is(err, ErrLoginfailed) || config.UpdateInterval == 0 || tries == 1 {
			log.Fatalf("Cannot connect to qbittorrent: %s. Exiting...", err)
		}
		tries--
		slog.Info("Cannot connect to qbittorrent", "error", err, "retrying in", loginDelay/time.Second, "remaining attempts", tries)
		time.Sleep(loginDelay)
	}
}

func main() {
	config := loadConfig()
	client := getQbitClient(config)
	run(context.Background(), config, client, glueGetter{})
}
