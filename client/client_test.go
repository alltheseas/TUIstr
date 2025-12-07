package client

import (
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func TestParsePrivKeyHex(t *testing.T) {
	hexKey := "1111111111111111111111111111111111111111111111111111111111111111"
	wantPub, err := nostr.GetPublicKey(hexKey)
	if err != nil {
		t.Fatalf("failed to get pub key: %v", err)
	}

	sk, pub, err := parsePrivKey(hexKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sk != hexKey {
		t.Fatalf("expected sk %s, got %s", hexKey, sk)
	}
	if pub != wantPub {
		t.Fatalf("expected pub %s, got %s", wantPub, pub)
	}
}

func TestParsePrivKeyNsec(t *testing.T) {
	hexKey := "2222222222222222222222222222222222222222222222222222222222222222"
	nsec, err := nip19.EncodePrivateKey(hexKey)
	if err != nil {
		t.Fatalf("failed to encode nsec: %v", err)
	}

	sk, pub, err := parsePrivKey(nsec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sk != hexKey {
		t.Fatalf("expected hex key, got %s", sk)
	}

	wantPub, _ := nostr.GetPublicKey(hexKey)
	if pub != wantPub {
		t.Fatalf("expected pub %s, got %s", wantPub, pub)
	}
}
