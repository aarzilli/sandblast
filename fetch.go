package sandblast

import (
	"net/http"
	"io/ioutil"
	"code.google.com/p/go.net/html/charset"
	"code.google.com/p/go.text/transform"
)

// Returns the body of resp as a decoded string, detecting its encoding
func DecodedBody(resp *http.Response) (content []byte, encoding string, err error) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		content = body
		return
	}
	e, encoding, _ := charset.DetermineEncoding(body, resp.Header.Get("Content-Type"))
	t := e.NewDecoder()
	content = make([]byte, len(body))
	start := 0
	for {
		var nDst, nSrc int
		nDst, nSrc, err = t.Transform(content[start:], body, true)
		body = body[nSrc:]
		start += nDst
		switch err {
		case transform.ErrShortDst:
			newContent := make([]byte, len(content)*2)
			copy(newContent, content)
			content = newContent
		case transform.ErrShortSrc:
			return
		default:
			content = content[:start]
			return
		}
	}
	return
}

func FetchURL(url string) (body []byte, status int, encoding string, err error) {
	resp, err := http.Get(url)
	if resp != nil {
		status = resp.StatusCode
	}
	if err != nil {
		return
	}
	body, encoding, err = DecodedBody(resp)
	return
}
