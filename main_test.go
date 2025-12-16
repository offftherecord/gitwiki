package main

import (
	"context"
	"testing"
)

func TestParseAccountInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedName string
	}{
		{
			name:         "org prefix",
			input:        "org:UrbanCompass",
			expectedType: "org",
			expectedName: "UrbanCompass",
		},
		{
			name:         "user prefix",
			input:        "user:offftherecord",
			expectedType: "user",
			expectedName: "offftherecord",
		},
		{
			name:         "no prefix",
			input:        "UrbanCompass",
			expectedType: "unknown",
			expectedName: "UrbanCompass",
		},
		{
			name:         "empty string",
			input:        "",
			expectedType: "unknown",
			expectedName: "",
		},
		{
			name:         "org prefix with colon in name",
			input:        "org:some:name",
			expectedType: "org",
			expectedName: "some:name",
		},
		{
			name:         "user prefix with colon in name",
			input:        "user:some:name",
			expectedType: "user",
			expectedName: "some:name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountType, accountName := parseAccountInput(tt.input)
			if accountType != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, accountType)
			}
			if accountName != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, accountName)
			}
		})
	}
}

func TestGetGitHubClient(t *testing.T) {
	ctx := context.Background()

	t.Run("creates client without token", func(t *testing.T) {
		client := getGitHubClient(ctx)
		if client == nil {
			t.Error("expected non-nil client")
		}
	})

	t.Run("creates client with token", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "test_token")
		client := getGitHubClient(ctx)
		if client == nil {
			t.Error("expected non-nil client")
		}
	})
}

func TestCheckWiki(t *testing.T) {
	tests := []struct {
		name string
		repo Repository
	}{
		{
			name: "repo without wiki",
			repo: Repository{
				Name:     "test-repo",
				URL:      "https://github.com/test/repo",
				HasWiki:  false,
				IsPublic: true,
			},
		},
		{
			name: "repo with invalid URL",
			repo: Repository{
				Name:     "test-repo",
				URL:      "://invalid-url",
				HasWiki:  true,
				IsPublic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkWiki(tt.repo)
		})
	}
}
