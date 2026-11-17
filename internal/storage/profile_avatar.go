package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
)

const (
	maxLocalAvatarBytes  = 8 << 20
	maxLocalAvatarPixels = 4_000_000
)

// LocalAvatarStore stores validated images by their SHA-256 digest. The digest
// keeps persisted preferences independent from user-selected source paths.
type LocalAvatarStore struct{ Root string }

func (s LocalAvatarStore) Save(data []byte) (string, error) {
	if _, err := decodeLocalAvatar(data); err != nil {
		return "", err
	}
	if s.Root == "" {
		return "", errors.New("avatar storage directory is empty")
	}

	digest := sha256.Sum256(data)
	key := hex.EncodeToString(digest[:])
	directory := filepath.Join(s.Root, "avatars")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(directory, key+".image")
	if existing, err := readAvatarFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return key, nil
		}
		if sha256.Sum256(existing) == digest {
			return "", errors.New("avatar SHA-256 collision")
		}
		if err := os.Remove(path); err != nil {
			return "", fmt.Errorf("remove corrupt avatar: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		if removeErr := os.Remove(path); removeErr != nil {
			return "", fmt.Errorf("remove unreadable avatar: %w", removeErr)
		}
	}

	temporary, err := os.CreateTemp(directory, ".avatar-*")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0600); err != nil {
		temporary.Close()
		return "", err
	}
	if count, err := temporary.Write(data); err != nil {
		temporary.Close()
		return "", err
	} else if count != len(data) {
		temporary.Close()
		return "", io.ErrShortWrite
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if existing, readErr := readAvatarFile(path); readErr == nil && bytes.Equal(existing, data) {
			return key, nil
		}
		return "", fmt.Errorf("activate local avatar: %w", err)
	}
	return key, nil
}

func (s LocalAvatarStore) Load(digest string) (image.Image, error) {
	if !validAvatarDigest(digest) || digest == "" {
		return nil, errors.New("invalid avatar digest")
	}
	if s.Root == "" {
		return nil, errors.New("avatar storage directory is empty")
	}
	data, err := readAvatarFile(filepath.Join(s.Root, "avatars", digest+".image"))
	if err != nil {
		return nil, err
	}
	actual := sha256.Sum256(data)
	if hex.EncodeToString(actual[:]) != digest {
		return nil, errors.New("stored avatar failed its SHA-256 check")
	}
	return decodeLocalAvatar(data)
}

func readAvatarFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxLocalAvatarBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxLocalAvatarBytes {
		return nil, errors.New("stored avatar exceeds 8 MB")
	}
	return data, nil
}

func decodeLocalAvatar(data []byte) (image.Image, error) {
	if len(data) == 0 || len(data) > maxLocalAvatarBytes {
		return nil, errors.New("avatar must be between 1 byte and 8 MB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("avatar must be a valid PNG, JPEG, or GIF image")
	}
	if format != "png" && format != "jpeg" && format != "gif" {
		return nil, errors.New("avatar must be a PNG, JPEG, or GIF image")
	}
	if config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > maxLocalAvatarPixels {
		return nil, errors.New("avatar image dimensions exceed the 4 megapixel limit")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("avatar image is incomplete or damaged")
	}
	return decoded, nil
}
