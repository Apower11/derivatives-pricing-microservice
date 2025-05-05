package posts

import(
	"bytes"
	"io"
	"net/http"
)

func getMIMEType(data []byte) string {
	buffer := bytes.NewReader(data)
	peekBytes := make([]byte, 512)
	_, err := buffer.Read(peekBytes)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}
	contentType := http.DetectContentType(peekBytes)
	return contentType
}