package grpc

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const minimumSharedTokenLength = 32

func loadOrCreateSharedToken() (string, error) {
	if token := strings.TrimSpace(os.Getenv("BRIDGE_CEP_TOKEN")); token != "" {
		return validateSharedToken(token, "BRIDGE_CEP_TOKEN")
	}
	if token := strings.TrimSpace(os.Getenv("MCP_CEP_TOKEN")); token != "" {
		return validateSharedToken(token, "MCP_CEP_TOKEN")
	}

	tokenPath := os.Getenv("PREMIERE_MCP_TOKEN_FILE")
	usingDefaultDirectory := tokenPath == ""
	if tokenPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory for bridge token: %w", err)
		}
		tokenPath = filepath.Join(home, ".premierpro-mcp", "cep-token")
	}

	if token, err := readSharedToken(tokenPath); err == nil {
		return token, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	tokenDir := filepath.Dir(tokenPath)
	_, statErr := os.Stat(tokenDir)
	directoryExisted := statErr == nil
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		return "", fmt.Errorf("create bridge token directory: %w", err)
	}
	if usingDefaultDirectory || !directoryExisted {
		if err := os.Chmod(tokenDir, 0o700); err != nil {
			return "", fmt.Errorf("secure bridge token directory: %w", err)
		}
	}

	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate bridge token: %w", err)
	}
	token := hex.EncodeToString(random)
	file, err := os.OpenFile(tokenPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return readSharedToken(tokenPath)
		}
		return "", fmt.Errorf("create bridge token file: %w", err)
	}
	if _, err := file.WriteString(token + "\n"); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write bridge token file: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close bridge token file: %w", err)
	}
	return token, nil
}

func readSharedToken(tokenPath string) (string, error) {
	contents, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	if err := os.Chmod(tokenPath, 0o600); err != nil {
		return "", fmt.Errorf("secure bridge token file: %w", err)
	}
	return validateSharedToken(strings.TrimSpace(string(contents)), tokenPath)
}

func validateSharedToken(token, source string) (string, error) {
	if len(token) < minimumSharedTokenLength {
		return "", fmt.Errorf("%s must contain at least %d characters", source, minimumSharedTokenLength)
	}
	return token, nil
}
