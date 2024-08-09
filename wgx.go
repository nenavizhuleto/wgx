package wgx

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	Exec   = "wg"
	QExec  = "wg-quick"
	TmpDir = "/tmp"
)

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

func Sanitize[T ~string](output T) T {
	return T(strings.Trim(string(output), "\n"))
}

type PrivateKey string

func (pk PrivateKey) Bytes() []byte {
	return []byte(pk)
}

func GenKey() (PrivateKey, error) {
	key, err := exec.Command(Exec, "genkey").Output()
	if err != nil {
		return "", err
	}

	return Sanitize(PrivateKey(key)), nil
}

type PublicKey string

func PubKey(pk PrivateKey) (PublicKey, error) {
	cmd := exec.Command(Exec, "pubkey")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	bytes := pk.Bytes()

	n, err := stdin.Write(bytes)
	if err != nil {
		return "", err
	}

	if n < len(bytes) {
		return "", io.ErrShortWrite
	}

	stdin.Close()

	key, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return Sanitize(PublicKey(key)), nil
}

func GenerateKeyPair() (KeyPair, error) {
	privateKey, err := GenKey()
	if err != nil {
		return KeyPair{}, err
	}

	publicKey, err := PubKey(privateKey)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, err
}

type Interface = string

func QStrip(iface Interface) ([]byte, error) {
	stripped, err := exec.Command(QExec, "strip", iface).Output()
	if err != nil {
		return nil, err
	}

	return stripped, nil
}

func SyncConf(iface Interface) error {
	stripped, err := QStrip(iface)
	if err != nil {
		return err
	}

	filename := path.Join(TmpDir, "syncconf.tmp")

	if err := os.WriteFile(filename, stripped, 0666); err != nil {
		return err
	}

	if err := exec.Command(Exec, "syncconf", iface, filename).Run(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	return nil
}
