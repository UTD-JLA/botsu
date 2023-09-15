package vn

import (
	"archive/tar"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/klauspost/compress/zstd"
)

var ErrUnknownTarType = errors.New("unknown tar type")

const VNDBDumpURL = "https://dl.vndb.org/dump/vndb-db-latest.tar.zst"

func DownloadVNDBDump(path string) (err error) {
	if err = os.MkdirAll(path, 0755); err != nil {
		return
	}

	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}

	resp, err := client.Get(VNDBDumpURL)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	decoder, err := zstd.NewReader(resp.Body)

	if err != nil {
		return
	}

	defer decoder.Close()

	reader := tar.NewReader(decoder)

	for {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			file, err := os.Create(path + "/" + header.Name)

			if err != nil {
				return err
			}

			defer file.Close()

			_, err = io.Copy(file, reader)

			if err != nil {
				return err
			}

		case tar.TypeDir:
			err := os.MkdirAll(path+"/"+header.Name, 0755)
			if err != nil {
				return err
			}
		default:
			return ErrUnknownTarType
		}
	}

	return
}
