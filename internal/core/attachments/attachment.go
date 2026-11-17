// Package attachments loads bounded local files for the offline chat preview.
package attachments

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
)

const MaxSize = 8 << 20

type Attachment struct {
	Name    string
	Data    []byte
	Preview image.Image
}

func Read(name string, r io.Reader) (*Attachment, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxSize {
		return nil, errors.New("file exceeds 8 MB")
	}
	name = filepath.Base(name)
	if name == "." || name == "" {
		name = "Attachment"
	}
	file := &Attachment{Name: name, Data: data}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil && config.Width > 0 && config.Height > 0 && int64(config.Width)*int64(config.Height) <= 4_000_000 {
		file.Preview, _, _ = image.Decode(bytes.NewReader(data))
	}
	return file, nil
}
