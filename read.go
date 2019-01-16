package readability

import (
	"math"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (self *TReadability) fixImagesPath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("img").Each(
		func(i int, img *goquery.Selection) {
			src, _ := img.Attr("src")
			// dz论坛的有些img属性使用的是file字段
			if f, ok := img.Attr("file"); ok {
				src = f
				img.SetAttr("src", f)
				img.RemoveAttr("file")
			}
			if f, ok := img.Attr("data-src"); ok {
				src = f
			}
			if src != "" && !strings.Contains(src, "data:image") {
				src = tr.fixLink(src)
				self.ImageList = append(self.ImageList, src)
				img.SetAttr("src", src)
			} else {
				img.Remove()
			}
		})
}
func (tr *TReadability) fixHrefPath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("a").Each(func(i int, link *goquery.Selection) {
		src, _ := link.Attr("href")
		src = tr.fixLink(src)
		if src == "" {
			link.Remove()
			return
		}
		link.SetAttr("href", src)
	})
}

func (self *TReadability) cleanConditionally(e *goquery.Selection, tag string) {
	if e == nil {
		return
	}
	contentScore := 0.0
	e.Find(tag).Each(func(i int, node *goquery.Selection) {
		weight := self.getClassWeight(node)
		hashNode := HashStr(node)
		if v, ok := self.candidates[hashNode]; ok {
			contentScore = v.score
		} else {
			contentScore = 0
		}

		if weight+contentScore < 0 {
			node.Remove()
		} else {
			p := node.Find("p").Length()
			img := node.Find("img").Length()
			li := node.Find("li").Length() - 100
			input_html := node.Find("input_html").Length()
			embedCount := 0
			node.Find("embed").Each(func(i int, embed *goquery.Selection) {
				if !videos.MatchString(embed.AttrOr("src", "")) {
					embedCount += 1
				}
			})
			linkDensity := self.getLinkDensity(node)
			contentLength := strLen(node.Text())
			toRemove := false
			if img > p && img > 1 {
				toRemove = true
			} else if li > p && tag != "ul" && tag != "ol" {
				toRemove = true
			} else if input_html > int(math.Floor(float64(p/3))) {
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

func (self *TReadability) cleanStyle(e *goquery.Selection) {
	if e == nil {
		return
	}
	e.Find("*").Each(func(i int, elem *goquery.Selection) {
		elem.RemoveAttr("class")
		elem.RemoveAttr("id")
		// elem.RemoveAttr("style")
		elem.RemoveAttr("width")
		elem.RemoveAttr("height")
		elem.RemoveAttr("onclick")
		elem.RemoveAttr("onmouseover")
		elem.RemoveAttr("border")
	})
}

func (self *TReadability) clean(e *goquery.Selection, tag string) {
	if e == nil {
		return
	}
	isEmbed := false
	if tag == "object" || tag == "embed" {
		isEmbed = true
	}
	e.Find(tag).Each(func(i int, target *goquery.Selection) {
		attributeValues := ""
		// TODO match v.qq.com
		if isEmbed && videos.MatchString(attributeValues) {
			return
		}
		if isEmbed && videos.MatchString(target.Text()) {
			return
		}
		target.Remove()
	})
}

func (self *TReadability) cleanArticle(content *goquery.Selection) {
	if content == nil {
		return ""
	}
	self.cleanStyle(content)
	self.clean(content, "h1")
	self.clean(content, "object")
	self.cleanConditionally(content, "form")
	if content.Find("h2").Length() == 1 {
		self.clean(content, "h2")
	}
	if content.Find("h3").Length() == 1 {
		self.clean(content, "h3")
	}
	self.clean(content, "iframe")
	self.cleanConditionally(content, "table")
	self.cleanConditionally(content, "ul")
	self.cleanConditionally(content, "div")

	self.fixImagesPath(content)
	self.fixHrefPath(content)

	summary := ""
	content.Find("p").Each(func(i int, s *goquery.Selection) {
		summary = summary + s.Text()
	})
	self.Summary = summary
	html, err := content.Html()
	if err != nil {
		return ""
	}
	// html = ghtml.UnescapeString(html)
	return killBreaks.ReplaceAllString(html, "<br />")
}
