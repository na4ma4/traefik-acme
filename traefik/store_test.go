package traefik_test

import (
	"slices"
	"testing"

	"github.com/na4ma4/traefik-acme/traefik"
)

func TestStore_FindTestExampleCom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data *[]byte
	}{
		{"traefik v1", &acmeDatav1},
		{"traefik v2", &acmeDatav2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store, err := traefik.ReadBytes(*tt.data, "acme")
			if err != nil {
				t.Fatalf("ReadBytes failed: %v", err)
			}
			if store == nil {
				t.Fatal("store is nil")
			}

			cert := store.GetCertificateByName("test.example.com")
			if cert == nil {
				t.Fatal("cert is nil")
			}
			if !slices.Contains(cert.Domain.ToStrArray(), "test.example.com") {
				t.Errorf("expected domains to contain test.example.com, got %v", cert.Domain.ToStrArray())
			}
		})
	}
}

func TestStore_FindAnotherTestExampleCom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data *[]byte
	}{
		{"traefik v1", &acmeDatav1},
		{"traefik v2", &acmeDatav2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store, err := traefik.ReadBytes(*tt.data, "acme")
			if err != nil {
				t.Fatalf("ReadBytes failed: %v", err)
			}
			if store == nil {
				t.Fatal("store is nil")
			}

			cert := store.GetCertificateByName("another-test.example.com")
			if cert == nil {
				t.Fatal("cert is nil")
			}
			if !slices.Contains(cert.Domain.ToStrArray(), "another-test.example.com") {
				t.Errorf("expected domains to contain another-test.example.com, got %v", cert.Domain.ToStrArray())
			}
		})
	}
}

func TestStore_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data *[]byte
	}{
		{"traefik v1", &acmeDatav1},
		{"traefik v2", &acmeDatav2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store, err := traefik.ReadBytes(*tt.data, "acme")
			if err != nil {
				t.Fatalf("ReadBytes failed: %v", err)
			}
			if store == nil {
				t.Fatal("store is nil")
			}

			cert := store.GetCertificateByName("test2.example.com")
			if cert != nil {
				t.Error("expected cert to be nil, got non-nil")
			}
		})
	}
}

func TestStore_CorruptData(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes([]byte("blah"), "acme")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if store != nil {
		t.Error("expected store to be nil, got non-nil")
	}
}

func TestStore_EmptyAcme(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes([]byte(`{"acme":{}}`), "acme")
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}
	if certs := store.GetCertificates(); len(certs) != 0 {
		t.Errorf("expected empty certificates, got %v", certs)
	}
}

func TestStore_ResolverNotFound(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes([]byte(`{}`), "acme")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if store != nil {
		t.Error("expected store to be nil, got non-nil")
	}
}

func TestStore_WildcardInSans(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes(acmeDatav3, "acme")
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}

	cert := store.GetCertificateByName("*.example.com")
	if cert == nil {
		t.Fatal("cert is nil")
	}
	if string(cert.Certificate) != "certificate-for-example.com\n" {
		t.Errorf("expected certificate %q, got %q", "certificate-for-example.com\n", string(cert.Certificate))
	}
	if string(cert.Key) != "key-for-example.com\n" {
		t.Errorf("expected key %q, got %q", "key-for-example.com\n", string(cert.Key))
	}
}

func TestStore_WildcardInMain(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes(acmeDatav4, "acme")
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}

	cert := store.GetCertificateByName("*.example.com")
	if cert == nil {
		t.Fatal("cert is nil")
	}
	if string(cert.Certificate) != "certificate-for-example.com\n" {
		t.Errorf("expected certificate %q, got %q", "certificate-for-example.com\n", string(cert.Certificate))
	}
	if string(cert.Key) != "key-for-example.com\n" {
		t.Errorf("expected key %q, got %q", "key-for-example.com\n", string(cert.Key))
	}
}

func TestStore_DifferentResolver(t *testing.T) {
	t.Parallel()

	store, err := traefik.ReadBytes(acmeDatav5, "acme-different")
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}

	cert := store.GetCertificateByName("example.com")
	if cert == nil {
		t.Fatal("cert is nil")
	}
	if string(cert.Certificate) != "certificate-for-example.com\n" {
		t.Errorf("expected certificate %q, got %q", "certificate-for-example.com\n", string(cert.Certificate))
	}
	if string(cert.Key) != "key-for-example.com\n" {
		t.Errorf("expected key %q, got %q", "key-for-example.com\n", string(cert.Key))
	}
}
