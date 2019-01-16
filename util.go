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
