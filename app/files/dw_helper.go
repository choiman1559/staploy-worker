package files

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

func DownloadFileWithUrl(url string, filepath string, headers map[string]string, config *tls.Config) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			return
		}
	}(out)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if config != nil {
		client.Transport = &http.Transport{TLSClientConfig: config}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

//goland:noinspection GoUnusedExportedFunction
func DownloadFileWithUrlMultipart(url string, savePath string, headers map[string]string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	contentType := resp.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return fmt.Errorf("not a multipart response")
	}

	mr := multipart.NewReader(resp.Body, params["boundary"])

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = func() error {
			defer func(part *multipart.Part) {
				err := part.Close()
				if err != nil {
					return
				}
			}(part)

			if part.FileName() != "" {
				out, err := os.Create(savePath)
				if err != nil {
					return err
				}
				defer func(out *os.File) {
					err := out.Close()
					if err != nil {
						return
					}
				}(out)

				_, err = io.Copy(out, part)
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("no file found in multipart response")
}
