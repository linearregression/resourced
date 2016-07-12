package writers

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/sethgrid/pester"
)

func init() {
	Register("Http", NewHttp)
}

// NewHttp is Http constructor.
func NewHttp() IWriter {
	httpWriter := &Http{}
	httpWriter.MaxRetries = 3 // Default

	return httpWriter
}

// Http is a writer that simply serialize all readers data to Http.
type Http struct {
	Base
	Url        string
	Method     string
	Headers    string
	Username   string
	Password   string
	MaxRetries int
}

// headersAsMap parses the headers data as string and returns them as map.
func (h *Http) headersAsMap() map[string]string {
	if h.Headers == "" {
		return nil
	}

	headersInMap := make(map[string]string)

	pairs := strings.Split(h.Headers, ",")

	for _, pairInString := range pairs {
		pair := strings.Split(pairInString, "=")
		if len(pair) >= 2 {
			headersInMap[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
		}
	}

	return headersInMap
}

// NewHttpRequest builds and returns http.Request struct.
func (h *Http) NewHttpRequest(dataJson []byte) (*http.Request, error) {
	var err error

	if h.Url == "" {
		return nil, errors.New("Url is undefined.")
	}

	if h.Method == "" {
		return nil, errors.New("Method is undefined.")
	}

	req, err := http.NewRequest(h.Method, h.Url, bytes.NewBuffer(dataJson))
	if err != nil {
		return nil, err
	}

	for key, value := range h.headersAsMap() {
		req.Header.Set(key, value)
	}

	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}

	return req, err
}

// Run executes the writer.
func (h *Http) Run() error {
	if h.Data == nil {
		return errors.New("Data field is nil.")
	}

	dataJson, err := h.ToJson()
	if err != nil {
		return err
	}

	req, err := h.NewHttpRequest(dataJson)
	if err != nil {
		return err
	}

	client := pester.New()
	client.MaxRetries = h.MaxRetries
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = false

	resp, err := client.Do(req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error":      err.Error(),
			"req.URL":    req.URL.String(),
			"req.Method": req.Method,
		}).Error("Failed to send HTTP request")

		return err
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	return nil
}
