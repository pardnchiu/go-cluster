package node

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pardnchiu/pdcluster/internal/util"
	"golang.org/x/crypto/ssh"
)

func (n *Node) KeyPair() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
	}

	folderPath := filepath.Join(home, ".ssh")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, 0700); err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
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
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
		}

		// * not passphrase, for eazy use in automation
		priKey, err := ssh.MarshalPrivateKey(pri, "")
		if err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
		}
		if err := os.WriteFile(keyPath, pem.EncodeToMemory(priKey), 0600); err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = AppName
		}

		// * generate public key
		pubKey, err := ssh.NewPublicKey(pub)
		if err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
		}
		pubBytes := ssh.MarshalAuthorizedKey(pubKey)
		pubLine := string(pubBytes[:len(pubBytes)-1]) + " root@" + hostname + "\n"
		if err := os.WriteFile(pubPath, []byte(pubLine), 0644); err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
		}

	}
	return nil
}
