package files

import (
	"io"
	"os"
)

func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ReadFileString(path string) (string, error) {
	content, err := ReadFile(path)
	return string(content), err
}

func WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0666)
}

func WriteFileString(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0666)
}

func SetExecutable(path string, executable bool) error {
	perm := 0666
	if executable {
		perm = 0777
	}
	return os.Chmod(path, os.FileMode(perm))
}

func CopyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {

		}
	}(srcFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {

		}
	}(destFile)

	buf := make([]byte, dirConfig.BufferSize)
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destFile.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}
