package codec

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

func WriteJson(filename string, object any) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(filename, data, os.ModePerm) //nolint:gosec
}

func WriteJsonZip(filename string, object any) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return err
	}

	return file.Sync()
}
