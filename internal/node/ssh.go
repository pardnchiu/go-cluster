package node

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func InitKeyPair() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed: no home folder",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	folderPath := filepath.Join(home, ".ssh")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, 0700); err != nil {
			slog.Error("failed: create folder",
				slog.String("path", folderPath),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}

	keyPath := filepath.Join(folderPath, "id_ed25519")
	pubPath := keyPath + ".pub"

	_, priErr := os.Stat(keyPath)
	_, pubErr := os.Stat(pubPath)

	if os.IsNotExist(priErr) || os.IsNotExist(pubErr) {
		// * generate private key
		pub, pri, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			slog.Error("failed: generate key pair",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		// * not passphrase, for eazy use in automation
		priKey, err := ssh.MarshalPrivateKey(pri, "")
		if err != nil {
			slog.Error("failed: marshal private key",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
		if err := os.WriteFile(keyPath, pem.EncodeToMemory(priKey), 0600); err != nil {
			slog.Error("failed: write private key",
				slog.String("path", keyPath),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "go-cluster"
		}

		// * generate public key
		pubKey, err := ssh.NewPublicKey(pub)
		if err != nil {
			slog.Error("failed: create public key",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
		pubBytes := ssh.MarshalAuthorizedKey(pubKey)
		pubLine := string(pubBytes[:len(pubBytes)-1]) + " root@" + hostname + "\n"
		if err := os.WriteFile(pubPath, []byte(pubLine), 0644); err != nil {
			slog.Error("failed: write public key",
				slog.String("path", pubPath),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		slog.Info("SSH key pair generated",
			slog.String("private", keyPath),
			slog.String("public", pubPath),
		)
	}
}
