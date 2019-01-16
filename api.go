package readability

import (
	"errors"
	nurl "net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TCandidateItem struct {
	score float64
	node  *goquery.Selection
}

type TReadability struct {
	html       string
	url        *nurl.URL
	htmlDoc    *goquery.Document
	candidates map[string]TCandidateItem

	Title   string
	Content string
	Cover   string

	Summary   string // 纯文字正文
	ImageList []string
}

func NewFromReader(content string, url string) (*TReadability, error) {

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

	tr.html = replaceBrs.ReplaceAllString(tr.html, "</p><p>")
	tr.html = strings.Replace(tr.html, "<noscript>", "", -1)
	tr.html = strings.Replace(tr.html, "</noscript>", "", -1)

	if tr.html == "" {
		return nil, errors.New("empty html")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(tr.html))
	if err != nil {
		return nil, err
	}

	tr.htmlDoc = doc

	tr.parse()

	return tr, nil
}

func (tr *TReadability) parse() {
	tr.Title = tr.htmlDoc.Find("title").Text()

	// start Parse body
	tr.preProcess()
	// extract the article
	if bodyNode := tr.extract(); bodyNode != nil {
		tr.cleanArticle(bodyNode)
	}
}

func (tr *TReadability) preProcess() {
	tr.htmlDoc.Find("script").Remove()
	tr.htmlDoc.Find("style").Remove()
	tr.htmlDoc.Find("link").Remove()

	tr.htmlDoc.Find("*").Each(func(i int, elem *goquery.Selection) {
		if tr.isComment(elem) {
			elem.Remove()
			return
		}
		unlikelyMatchString := elem.AttrOr("id", "") + " " + elem.AttrOr("class", "")

		if unlikelyCandidates.MatchString(unlikelyMatchString) &&
			!okMaybeItsACandidate.MatchString(unlikelyMatchString) &&
			tr.getTagName(elem) != "body" {
			elem.Remove()
			return
		}
		if unlikelyElements.MatchString(tr.getTagName(elem)) {
			elem.Remove()
			return
		}
		if tr.getTagName(elem) == "div" {
			s, _ := elem.Html()
			if !divToPElements.MatchString(s) {
				elem.Nodes[0].Data = "p"
			}
		}
	})
}
