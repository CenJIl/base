package web

import (
	"mime/multipart"
	"path/filepath"
	"testing"
)

func TestValidateFileAllowsUnlimitedSizeWhenUnset(t *testing.T) {
	file := &multipart.FileHeader{Filename: "photo.JPG", Size: 1024}
	if err := ValidateFile(file, UploadConfig{AllowedExts: []string{".jpg"}}); err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
}

func TestValidateFileRejectsNilFile(t *testing.T) {
	if err := ValidateFile(nil, UploadConfig{}); err == nil {
		t.Fatal("ValidateFile(nil) error = nil, want error")
	}
}

func TestDefaultCORSDoesNotAllowWildcardOrigins(t *testing.T) {
	cfg := (Config{}).DefaultCORS()
	for _, origin := range cfg.AllowOrigins {
		if origin == "*" {
			t.Fatal("default CORS must not allow wildcard origins")
		}
	}
}

func TestDefaultCORSDisablesCredentialsForWildcardOrigin(t *testing.T) {
	cfg := (Config{CORS: CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}}).DefaultCORS()
	if cfg.AllowCredentials {
		t.Fatal("credentialed wildcard CORS must be disabled")
	}
}

type webTestConfig struct{ Config }

func TestNewServerEReturnsMissingConfigError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	h, err := NewServerE[webTestConfig](path)
	if err == nil {
		t.Fatal("NewServerE() error = nil, want missing configuration error")
	}
	if h != nil {
		t.Fatalf("NewServerE() server = %#v, want nil", h)
	}
}

func TestNewIPRateLimiterNormalizesInvalidLimit(t *testing.T) {
	limiter := NewIPRateLimiter(0, 0)
	t.Cleanup(limiter.Close)
	if !limiter.Allow("127.0.0.1") {
		t.Fatal("normalized rate limiter must allow its initial request")
	}
}
