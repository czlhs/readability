package readability

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	nurl "net/url"

	shtml "html"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	unlikelyCandidates   = regexp.MustCompile(`(?i)combx|modal|comment|disqus|foot|header|menu|meta|nav|rss|shoutbox|sidebar|sponsor|social|teaserlist|time|tweet|twitter`)
	okMaybeItsACandidate = regexp.MustCompile(`(?im)and|article|body|column|main|story|entry|^post`)
	positive             = regexp.MustCompile(`(?is)article|body|content|entry|hentry|main|page|pagination|post|text|blog|story`)
	negative             = regexp.MustCompile(`(?is)combx|comment|com|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tags|tool|widget`)
	extraneous           = regexp.MustCompile(`(?is)print|archive|comment|discuss|e[\-]?mail|share|reply|all|login|sign|single`)
	divToPElements       = regexp.MustCompile(`(?is)<(a|blockquote|dl|div|img|ol|p|pre|table|ul)`)
	replaceBrs           = regexp.MustCompile(`(?is)(<br[^>]*>[ \n\r\t]*){2,}`)
	replaceFonts         = regexp.MustCompile(`(?is)<(/?)font[^>]*>`)
	trim                 = regexp.MustCompile(`(?is)^\s+|\s+$`)
	normalize            = regexp.MustCompile(`(?is)\s{2,}`)
	killBreaks           = regexp.MustCompile(`(?is)(<br\s*/?>(\s|&nbsp;?)*)+`)
	videos               = regexp.MustCompile(`(?is)https?:\/\/(www\.|v\.)?(qq|youtube|vimeo|youku|tudou|56|yinyuetai)\.com`)
	videosInPath         = regexp.MustCompile(`(?is)\.com\/\w*video\w*`)
	attributeScore       = regexp.MustCompile(`(?is)blog|post|article`)
	skipFootnoteLink     = regexp.MustCompile(`(?is)^\s*(\[?[a-z0-9]{1,2}\]?|^|edit|citation needed)\s*$"`)
	nextLink             = regexp.MustCompile(`(?is)(next|weiter|continue|>([^\|]|$)|»([^\|]|$))`)
	prevLink             = regexp.MustCompile(`(?is)(prev|earl|old|new|<|«)`)

	unlikelyElements = regexp.MustCompile(`(?is)(input|time|button)`)

	contentMayContain = regexp.MustCompile(`\.( |$)`)
	pageCodeReg       = regexp.MustCompile(`(?is)<meta.+?charset=[^\w]?([-\w]+)`)
	validURL          = regexp.MustCompile(`^(https?)?://(www\.)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
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

	Title     string
	Content   string
	Summary   string
	ImageList []string
	Cover     string
}

func NewReadability(r io.Reader, url string) (v *TReadability, err error) {
	v = &TReadability{
		candidates: make(map[string]TCandidateItem, 0),
	}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	v.Cover = v.getCover(doc.Selection)
	v.Title = shtml.UnescapeString(v.getTitle(doc.Selection))

	v.prepareArticle(doc.Selection)
	articleDoc := v.grabArticle(doc.Selection)
	if articleDoc == nil {
		return
	}

	v.cleanArticle(articleDoc)
	base, _ := nurl.Parse(url)
	if base.Scheme == "" {
		base.Scheme = "http"
	}
	v.fixRelativeURIs(articleDoc, base)

	article, _ := articleDoc.Html()
	v.Content = killBreaks.ReplaceAllString(article, "<br />")
	v.Summary = shtml.UnescapeString(strings.TrimSpace(articleDoc.Text()))

	return v, nil
}

func (tr *TReadability) prepareArticle(doc *goquery.Selection) {
	doc.Find("script").Remove()
	doc.Find("noscript").Remove()
	doc.Find("link").Remove()
}

func (tr *TReadability) getCover(doc *goquery.Selection) string {
	var tmpCover, cover string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		s.Html()
		if prop := s.AttrOr("property", ""); prop == "og:image" {
			cover, _ = s.Attr("content")
		}
		s.Text()
		if itemprop := s.AttrOr("itemprop", ""); itemprop == "image" {
			tmpCover, _ = s.Attr("content")
		}
	})
	if cover == "" {
		cover = tmpCover
	}

	if validURL.MatchString(cover) {
		return toAbsoluteURI(cover, tr.url)
	}

	return ""
}

func (tr *TReadability) getTitle(doc *goquery.Selection) string {
	return strings.TrimSpace(doc.Find("title").Text())
}

func (tr *TReadability) fixRelativeURIs(doc *goquery.Selection, base *nurl.URL) {
	if doc == nil {
		return
	}
	doc.Find("img").Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		if dataSrc, _ := img.Attr("src"); dataSrc != "" {
			src = dataSrc
		}
		src = toAbsoluteURI(src, base)
		if src == "" {
			img.Remove()
		}
		img.SetAttr("src", src)
		tr.ImageList = append(tr.ImageList, src)
	})
	doc.Find("a").Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("href")
		src = toAbsoluteURI(src, base)
		if src == "" {
			img.Remove()
		}
		img.SetAttr("href", src)
	})
	doc.Find("video,iframe,embed,object").Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		src = toAbsoluteURI(src, base)
		if src == "" {
			img.Remove()
		}
		img.SetAttr("src", src)
	})
}

func (tr *TReadability) cleanConditionally(e *goquery.Selection, tag string) {
	if e == nil {
		return
	}
	e.Find(tag).Each(func(i int, node *goquery.Selection) {
		weight := tr.getClassWeight(node)

		if weight < 0 {
			node.Remove()
		} else {
			p := node.Find("p").Length()
			img := node.Find("img").Length()
			li := node.Find("li").Length() - 100
			input := node.Find("input").Length()

			embedCount := 0
			node.Find("embed").Each(func(i int, embed *goquery.Selection) {
				if !videos.MatchString(embed.AttrOr("src", "")) {
					embedCount++
				}
			})

			linkDensity := tr.getLinkDensity(node)
			contentLength := strLen(node.Text())
			toRemove := false
			if img > p && img > 1 {
				toRemove = true
			} else if li > p && tag != "ul" && tag != "ol" {
				toRemove = true
			} else if input > int(math.Floor(float64(p/3))) {
				toRemove = true
			} else if contentLength < 25 && (img == 0 || img > 2) {
				toRemove = true
			} else if weight < 25 && linkDensity > 0.2 {
				toRemove = true
			} else if weight >= 25 && linkDensity > 0.5 {
				toRemove = true
			} else if (embedCount == 1 && contentLength < 35) || embedCount > 1 {
				toRemove = true
			}
			if toRemove {
				node.Remove()
			}
		}
	})
}

func (tr *TReadability) cleanStyle(e *goquery.Selection) {
	if e == nil {
		return
	}
	e.Find("*").Each(func(i int, elem *goquery.Selection) {
		// elem.RemoveAttr("class")
		// elem.RemoveAttr("id")
		// elem.RemoveAttr("style")
		elem.RemoveAttr("width")
		elem.RemoveAttr("height")
		elem.RemoveAttr("onclick")
		elem.RemoveAttr("onmouseover")
		elem.RemoveAttr("border")
	})
}

func (tr *TReadability) clean(e *goquery.Selection, tag string) {
	if e == nil {
		return
	}
	isEmbed := false
	if tag == "object" || tag == "embed" || tag == "iframe" {
		isEmbed = true
	}
	e.Find(tag).Each(func(i int, target *goquery.Selection) {
		attributeValues, _ := target.Html()
		if isEmbed && (videos.MatchString(attributeValues) || videosInPath.MatchString(attributeValues)) {
			return
		}
		target.Remove()
	})
}

func (tr *TReadability) getTagName(node *goquery.Selection) string {
	if node == nil {
		return ""
	}
	return node.Nodes[0].Data
}

func (tr *TReadability) grabArticle(doc *goquery.Selection) *goquery.Selection {
	doc.Find("*").Each(func(i int, elem *goquery.Selection) {
		if elem.Nodes[0].Type == html.CommentNode {
			elem.Remove()
			return
		}
		unlikelyMatchString := elem.AttrOr("id", "") + "" + elem.AttrOr("class", "")
		if unlikelyCandidates.MatchString(unlikelyMatchString) &&
			!okMaybeItsACandidate.MatchString(unlikelyMatchString) &&
			tr.getTagName(elem) != "body" && tr.getTagName(elem) != "html" {
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
			} else {
				elem.Contents().Each(func(_ int, s *goquery.Selection) {
					if s.Nodes[0].Type == html.TextNode {
						nextSibling := s.Nodes[0].NextSibling
						if nextSibling != nil && nextSibling.Type == html.ElementNode && nextSibling.Data == "br" {
							s.Next().Remove() // 先删后面的再替换,否则带回s不指向br了
							s = s.ReplaceWithHtml(`<p>` + s.Text() + `</p>`)
						} else {
							s.ReplaceWithHtml(`<span>` + s.Text() + `</span>`)
						}
					}
				})
			}
		}
	})
	doc.Find("p").Each(func(i int, node *goquery.Selection) {
		parentNode := node.Parent()
		grandParentNode := parentNode.Parent()
		innerText := node.Text()

		if parentNode == nil || strLen(innerText) < 20 {
			return
		}
		parentHash := HashStr(parentNode)
		grandParentHash := HashStr(grandParentNode)
		if _, ok := tr.candidates[parentHash]; !ok {
			tr.candidates[parentHash] = tr.initializeNode(parentNode)
		}
		if _, ok := tr.candidates[grandParentHash]; !ok {
			tr.candidates[grandParentHash] = tr.initializeNode(grandParentNode)
		}
		contentScore := 1.0
		contentScore += float64(strings.Count(innerText, ","))
		contentScore += float64(strings.Count(innerText, "，"))
		contentScore += math.Min(math.Floor(float64(strLen(innerText)/100)), 3)

		v, _ := tr.candidates[parentHash]
		v.score += contentScore
		tr.candidates[parentHash] = v

		v, _ = tr.candidates[grandParentHash]
		v.score += contentScore / 2.0
		tr.candidates[grandParentHash] = v
	})

	delete(tr.candidates, "") // 删掉空节点

	var topCandidate *TCandidateItem
	for k, v := range tr.candidates {
		v.score = v.score * (1 - tr.getLinkDensity(v.node))
		tr.candidates[k] = v
		if topCandidate == nil || v.score > topCandidate.score {
			if topCandidate == nil {
				topCandidate = new(TCandidateItem)
			}
			topCandidate.score = v.score
			topCandidate.node = v.node
		}
	}
	if topCandidate != nil {
		siblingScoreThreshold := math.Max(10, topCandidate.score*0.2)
		topCandidate.node.Siblings().Each(func(_ int, s *goquery.Selection) {
			shouldAppend := false
			if v, ok := tr.candidates[HashStr(s)]; ok && v.score > siblingScoreThreshold {
				shouldAppend = true
			} else {
				text := s.Text()
				textLen := strLen(text)
				linkDensity := tr.getLinkDensity(s)
				if textLen > 80 && linkDensity < 0.25 {
					shouldAppend = true
				} else if textLen < 80 && linkDensity == 0.0 && contentMayContain.MatchString(text) {
					shouldAppend = true
				}
			}
			if shouldAppend {
				topCandidate.node = topCandidate.node.AppendSelection(s)
			}
		})
		return topCandidate.node
	}
	return nil
}

func (tr *TReadability) cleanArticle(content *goquery.Selection) *goquery.Selection {
	if content == nil {
		return nil
	}
	tr.cleanStyle(content)
	tr.clean(content, "object")
	tr.clean(content, "form")

	if content.Find("h1").Length() == 1 {
		tr.clean(content, "h2")
	}
	if content.Find("h2").Length() == 1 {
		tr.clean(content, "h2")
	}
	if content.Find("h3").Length() == 1 {
		tr.clean(content, "h3")
	}
	tr.clean(content, "iframe")
	tr.cleanConditionally(content, "table")
	tr.cleanConditionally(content, "ul")
	tr.cleanConditionally(content, "div")

	return content
}

func (tr *TReadability) getLinkDensity(node *goquery.Selection) float64 {
	if node == nil {
		return 0
	}
	textLength := float64(strLen(node.Text()))
	if textLength == 0 {
		return 0
	}
	linkLength := 0.0
	node.Find("a").Each(
		func(i int, link *goquery.Selection) {
			if href, _ := link.Attr("href"); href != "" && href[0] != '#' {
				linkLength += float64(strLen(link.Text()))
			}
		})
	return linkLength / textLength
}

func (tr *TReadability) getClassWeight(node *goquery.Selection) float64 {
	weight := 0.0
	if str, b := node.Attr("class"); b {
		if negative.MatchString(str) {
			weight -= 25
		}
		if positive.MatchString(str) {
			weight += 25
		}
	}
	if str, b := node.Attr("id"); b {
		if negative.MatchString(str) {
			weight -= 25
		}
		if positive.MatchString(str) {
			weight += 25
		}
	}
	return weight
}
func (tr *TReadability) initializeNode(node *goquery.Selection) TCandidateItem {
	contentScore := 0.0
	switch tr.getTagName(node) {
	case "article":
		contentScore += 10
	case "section":
		contentScore += 8
	case "div":
		contentScore += 5
	case "pre", "blockquote", "td":
		contentScore += 3
	case "form", "ol", "dl", "dd", "dt", "li", "address":
		contentScore -= 3
	case "th", "h1", "h2", "h3", "h4", "h5", "h6":
		contentScore -= 5
	}
	contentScore += tr.getClassWeight(node)
	return TCandidateItem{contentScore, node}
}

func HashStr(node *goquery.Selection) string {
	if node == nil || node.Length() == 0 {
		return ""
	}
	return fmt.Sprintf("%x", node.Nodes[0])
}

func strLen(str string) int {
	return utf8.RuneCountInString(str)
}

func toAbsoluteURI(url string, base *nurl.URL) string {
	if base == nil {
		return url
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(url, "//") {
			url = base.Scheme + ":" + url
		} else if strings.HasPrefix(url, "/") {
			url = base.Scheme + "://" + base.Host + url
		} else {
			url = base.Scheme + "://" + base.Host + base.Path + url
		}
		return url
	}

	return url
}
