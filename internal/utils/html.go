package utils

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func GetHtml(url string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return nil, errors.New("response body type is not HTML")
	}

	return io.ReadAll(resp.Body)
}

func ParseHtmlStreamWithCallback(htmlBytes []byte, callback func(*html.Node) string) (string, error) {
	node, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return "", err
	}
	result := traverse(node, callback)
	if len(result) == 0 {
		return "", errors.New(" value not found")
	}
	return result, nil
}

func FindLtTicket(node *html.Node) string {
	for _, val := range node.Attr {
		if val.Key == "name" && val.Val == "lt" {
			for _, v := range node.Attr {
				if v.Key == "value" {
					return v.Val
				}
			}
		}
	}
	return ""
}

func FindExecution(node *html.Node) string {
	var foundExecution bool
	for _, attr := range node.Attr {
		if attr.Key == "name" && attr.Val == "execution" {
			foundExecution = true
		}
		if attr.Key == "value" && foundExecution {
			return attr.Val
		}
	}
	return ""
}

func traverse(n *html.Node, callback func(n *html.Node) string) string {
	if n.Type == html.ElementNode {
		result := callback(n)
		if len(result) != 0 {
			return result
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := traverse(c, callback)
		if len(result) != 0 {
			return result
		}
	}
	return ""
}

func NewFormData(username, password, ltTicket, execution string) url.Values {
	return url.Values{
		"username":  {username},
		"password":  {password},
		"lt":        {ltTicket},
		"execution": {execution},
		"_eventId":  {"submit"},
		"warn":      {"false"},
		"submit":    {"Zaloguj"},
	}
}
