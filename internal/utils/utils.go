package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
)

func HTTPRequest(method, url string, body []byte, headers map[string]string, Username string, Password string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))

	fmt.Println(req)

	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if Username != "" && Password != "" {
		req.SetBasicAuth(Username, Password)
	}
	return http.DefaultClient.Do(req)

}

func BoolToStr(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func StringToInt(value string) int {

	// string to int
	i, err := strconv.Atoi(value)
	if err != nil {
		// ... handle error
		panic(err)
	}

	return i
}
