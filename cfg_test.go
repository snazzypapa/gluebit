package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test case 1: valid config with --gluetunhost and --gluetunport
	os.Args = []string{"cmd", "--qbituser", "user", "--qbitpass", "pass", "--qbithost", "localhost", "--qbitport", "8080", "--gluetunhost", "localhost", "--gluetunport", "8000", "--interval", "60"}
	config := loadConfig()
	if config.QbitUsername != "user" {
		t.Errorf("Expected QbitUsername to be 'user', but got '%s'", config.QbitUsername)
	}
	if config.QbitPassword != "pass" {
		t.Errorf("Expected QbitPassword to be 'pass', but got '%s'", config.QbitPassword)
	}
	if config.QbitHost != "localhost" {
		t.Errorf("Expected QbitHost to be 'localhost', but got '%s'", config.QbitHost)
	}
	if config.QbitPort != 8080 {
		t.Errorf("Expected QbitPort to be 8080, but got %d", config.QbitPort)
	}
	if config.GlueTunHost != "localhost" {
		t.Errorf("Expected GlueTunHost to be 'localhost', but got '%s'", config.GlueTunHost)
	}
	if config.GlueTunPort != 8000 {
		t.Errorf("Expected GlueTunPort to be 8000, but got %d", config.GlueTunPort)
	}
	if config.UpdateInterval != 60 {
		t.Errorf("Expected UpdateInterval to be 60, but got %d", config.UpdateInterval)
	}

	// Test case 2: valid config with --gluetunportfile
	os.Args = []string{"cmd", "--qbituser", "user", "--qbitpass", "pass", "--qbithost", "localhost", "--qbitport", "8080", "--gluetunportfile", "/path/to/portfile"}
	config = loadConfig()
	if config.QbitUsername != "user" {
		t.Errorf("Expected QbitUsername to be 'user', but got '%s'", config.QbitUsername)
	}
	if config.QbitPassword != "pass" {
		t.Errorf("Expected QbitPassword to be 'pass', but got '%s'", config.QbitPassword)
	}
	if config.QbitHost != "localhost" {
		t.Errorf("Expected QbitHost to be 'localhost', but got '%s'", config.QbitHost)
	}
	if config.QbitPort != 8080 {
		t.Errorf("Expected QbitPort to be 8080, but got %d", config.QbitPort)
	}
	if config.GlueTunPortFile != "/path/to/portfile" {
		t.Errorf("Expected GlueTunPortFile to be '/path/to/portfile', but got '%s'", config.GlueTunPortFile)
	}
	if config.UpdateInterval != 0 {
		t.Errorf("Expected UpdateInterval to be 0, but got %d", config.UpdateInterval)
	}
}
