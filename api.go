package readability

import (
	nurl "net/url"
	"strings"
)

type TCandidateItem struct {
	score float64
	node  *Node
}

type TReadability struct {
	html       string
	url        *nurl.URL
	candidates map[string]TCandidateItem
	root       *Node

	Title   string
	Content string
	Cover   string

	Summary   string // 纯文字正文
	ImageList []string
}

func NewFromHTML(content string, url string) (*TReadability, error) {

	tr := &TReadability{
		html:       content,
		candidates: make(map[string]TCandidateItem, 0),
		ImageList:  []string{},
	}

	if tr.url, _ = nurl.Parse(url); tr.url == nil {
		tr.url = &nurl.URL{
			Scheme: "http",
		}
	}

	// tr.html = replaceBrs.ReplaceAllString(tr.html, "</p><p>")
	// tr.html = strings.Replace(tr.html, "<noscript>", "", -1)
	// tr.html = strings.Replace(tr.html, "</noscript>", "", -1)
	doc, err := NewDocument(strings.NewReader(tr.html))
	// doc, err := goquery.NewDocumentFromReader(strings.NewReader(tr.html))
	if err != nil {
		return nil, err
	}
	tr.root = doc

	tr.parse()

	return tr, nil
}

func (tr *TReadability) parse() {
	// tr.Cover = tr.getCover()
	// tr.Title = tr.htmlDoc.Find("title").Text()
	// start Parse body
	tr.preProcess()
	// extract the article
	tr.grabArticle()
}

func (tr *TReadability) preProcess() {
	tr.root.WorkElementNode("script", func(n *Node) {
		n.Remove()
	})
	// tr.htmlDoc.Find("style").Remove()
	// tr.htmlDoc.Find("link").Remove()
}

// 需要在获取content之前调
func (tr *TReadability) getCover() string {
	// var tmpCover, cover string
	// tr.htmlDoc.Find("meta").Each(func(i int, s *goquery.Selection) {
	// 	s.Html()
	// 	if prop, ok := s.Attr("property"); ok && prop == "og:image" {
	// 		cover, _ = s.Attr("content")
	// 	}
	// 	s.Text()
	// 	if itemprop, ok := s.Attr("itemprop"); ok && itemprop == "image" {
	// 		tmpCover, _ = s.Attr("content")
	// 	}
	// })
	// if cover == "" {
	// 	cover = tmpCover
	// }

	// if validURL.MatchString(cover) {
	// 	return tr.fixLink(cover)
	// }

	return ""
}
