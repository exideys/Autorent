package main

import (
	"reflect"
	"testing"
)

func TestCORSConfigDefaultsToWildcardWithoutCredentials(t *testing.T) {
	cfg := corsConfig("")

	if !reflect.DeepEqual(cfg.AllowOrigins, []string{"*"}) {
		t.Fatalf("unexpected origins: %+v", cfg.AllowOrigins)
	}
	if cfg.AllowCredentials {
		t.Fatal("wildcard origins must not allow credentials")
	}
}

func TestCORSConfigUsesConfiguredOriginsWithCredentials(t *testing.T) {
	cfg := corsConfig("https://example.com,https://admin.example.com")

	expectedOrigins := []string{"https://example.com", "https://admin.example.com"}
	if !reflect.DeepEqual(cfg.AllowOrigins, expectedOrigins) {
		t.Fatalf("expected origins %+v, got %+v", expectedOrigins, cfg.AllowOrigins)
	}
	if !cfg.AllowCredentials {
		t.Fatal("configured origins should allow credentials")
	}
	if !reflect.DeepEqual(cfg.AllowMethods, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}) {
		t.Fatalf("unexpected methods: %+v", cfg.AllowMethods)
	}
}
