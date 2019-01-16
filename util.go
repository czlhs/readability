package readability

import (
	"fmt"
	"strings"

	"crypto/md5"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func HashStr(node *goquery.Selection) string {
	if node == nil {
		return ""
	}
	html, _ := node.Html()
	return fmt.Sprintf("%x", md5.Sum([]byte(html)))
}

func strLen(str string) int {
	return utf8.RuneCountInString(str)
}
func (tr *TReadability) getTagName(node *goquery.Selection) string {
	if node == nil {
		return ""
	}
	return node.Nodes[0].Data
}

func (tr *TReadability) isComment(node *goquery.Selection) bool {
	if node == nil {
		return false
	}
	return node.Nodes[0].Type == html.CommentNode
}

func (tr *TReadability) fixLink(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(url, "//") {
			url = tr.url.Scheme + ":" + url
		} else if strings.HasPrefix(url, "/") {
			url = tr.url.Scheme + "://" + tr.url.Host + url
		} else {
			url = tr.url.Scheme + "://" + tr.url.Host + tr.url.Path + url
		}
		return url
	}

	return url
}

func CheckHide(node *goquery.Selection) bool {
	style, ok := node.Attr("style")
	if !ok {
		return false
	}

	kv := ParseStyle(style)
	if v, ok := kv["display"]; ok && v == "none" {
		return true
	}
	if v, ok := kv["visibility"]; ok && v == "hidden" {
		return true
	}
	if v, ok := kv["overflow"]; ok && v == "hidden" {
		return true
	}

	return false
}

func ParseStyle(style string) map[string]string {
	m := map[string]string{}
	styleAttrs := strings.Split(style, ";")
	for _, styleAttr := range styleAttrs {
		keyValue := strings.Split(styleAttr, ":")
		if len(keyValue) != 2 {
			continue
		}
		key := strings.TrimSpace(keyValue[0])
		m[key] = strings.TrimSpace(keyValue[1])
	}
	return m
}
