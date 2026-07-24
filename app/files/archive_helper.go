package files

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractTar(tarFile string, targetDir string) error {
	f, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)

	tr := tar.NewReader(f)

	destAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destAbs, header.Name)
		rel, err := filepath.Rel(destAbs, targetPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("wrong filepath (%s)", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := MkdirAll(targetPath); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := MkdirAll(filepath.Dir(targetPath)); err != nil {
				return err
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				err := outFile.Close()
				if err != nil {
					return err
				}
				return err
			}

			err = outFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
