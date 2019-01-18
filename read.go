package readability

import (
	"math"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	v "ixiaochuan.cn/connsrv/test"
)

func (tr *TReadability) grabArticle() {

	tr.root.Work(func(node *Node) {
		if node.IsComment() {
			node.Remove()
			return
		}
		unlikelyMatchString := node.GetAttrOr("id", "") + " " + node.GetAttrOr("class", "")

		if unlikelyCandidates.MatchString(unlikelyMatchString) &&
			!okMaybeItsACandidate.MatchString(unlikelyMatchString) &&
			node.GetTagName() != "body" {
			node.Remove()
			return
		}

		if unlikelyElements.MatchString(node.GetTagName()) {
			node.Remove()
			return
		}
		if node.GetTagName() == "div" {
			s, _ := node.Html()
			if !divToPElements.MatchString(s) {
				node.Data = "p"
			} else {
			}
		}
	})

	tr.root.WorkElementNode("p", func(node *Node) {
		parentNode := node.ParentNode()
		grandParentNode := parentNode.ParentNode()
		innerText := node.InnerText()

		if strLen(innerText) < 25 {
			return
		}
		contentScore := 1.0
		contentScore += float64(strings.Count(innerText, ","))
		contentScore += float64(strings.Count(innerText, "，"))
		contentScore += math.Min(math.Floor(float64(strLen(innerText)/100)), 3)

		if parentNode != nil {
			parentHash := HashStr(parentNode)
			if _, ok := tr.candidates[parentHash]; !ok {
				tr.candidates[parentHash] = tr.initializeNode(parentNode)
			}
			v, _ := tr.candidates[parentHash]
			v.score += contentScore
			tr.candidates[parentHash] = v
		}
		if grandParentNode != nil {
			grandParentHash := HashStr(grandParentNode)
			if _, ok := tr.candidates[grandParentHash]; !ok {
				tr.candidates[grandParentHash] = tr.initializeNode(grandParentNode)
			}
			v, _ = tr.candidates[grandParentHash]
			v.score += contentScore / 2.0
			tr.candidates[grandParentHash] = v
		}
	})

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
	if topCandidate == nil {
		return
	}
	articleDoc := NewNode(html.ElementNode)
	siblingScoreThreshold := math.Max(10, topCandidate.score*0.2)
	for c := topCandidate.node.Parent.FirstChild; c != nil; c = c.NextSibling {
		n := &Node{c}
		shouldAppend := false
		if c == topCandidate.node.Node {
			shouldAppend = true
		}

		if candidates, ok := tr.candidates[HashStr(n)]; ok && candidates.score > siblingScoreThreshold {
			shouldAppend = true
		}

		if n.GetTagName() == "p" {
			linkDensity := tr.getLinkDensity(n)
			text := n.InnerText()
			textLength := len([]rune(text))
			if textLength > 80 && linkDensity < 0.25 {
				shouldAppend = true
			} else if textLength < 80 && linkDensity == 0.0 && contentMayContain.MatchString(text) {
				shouldAppend = true
			} // TODO
		}
		if shouldAppend {
			articleDoc.AppendChild(n)
		}
	}
	tr.cleanArticle(articleDoc)
	return
}

func (tr *TReadability) fixImagesPath(node *goquery.Selection) {
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
				tr.ImageList = append(tr.ImageList, src)
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
func (tr *TReadability) fixIframePath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("iframe").Each(func(i int, link *goquery.Selection) {
		src, _ := link.Attr("src")
		src = tr.fixLink(src)
		if src == "" {
			link.Remove()
			return
		}
		link.SetAttr("src", src)
	})
}

func (tr *TReadability) fixEmbedPath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("embed").Each(func(i int, link *goquery.Selection) {
		src, _ := link.Attr("src")
		src = tr.fixLink(src)
		if src == "" {
			link.Remove()
			return
		}
		link.SetAttr("src", src)
	})
}
func (tr *TReadability) fixVideoPath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("video").Each(func(i int, video *goquery.Selection) {
		video.Find("source").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			if src = tr.fixLink(src); src == "" {
				s.Remove()
				return
			}
			s.SetAttr("src", src)
		})
	})
}

func (tr *TReadability) fixObjectPath(node *goquery.Selection) {
	if node == nil {
		return
	}
	node.Find("object").Each(func(i int, s *goquery.Selection) {
		data, _ := s.Attr("data")
		if data = tr.fixLink(data); data == "" {
			s.Remove()
			return
		}
		s.SetAttr("data", data)
	})
}

func (tr *TReadability) cleanConditionally(node *Node, tag string) {
	if node == nil {
		return
	}
	node.WorkElementNode(tag, func(n *Node) {
		weight := tr.getClassWeight(node)
		// hashNode := HashStr(node)
		// if v, ok := tr.candidates[hashNode]; ok {
		// 	contentScore = v.score
		// } else {
		// 	contentScore = 0
		// }

		if weight < 0 {
			node.Remove()
		} else if len(strings.Split(node.InnerText(), ",")) < 10 {
			p := node.TagLength("p")
			img := node.TagLength("img")
			li := node.TagLength("li") - 100
			input := node.TagLength("input")

			embedCount := 0
			node.WorkElementNode("embed", func(n *Node) {
				if !videos.MatchString(n.GetAttrOr("src", "")) {
					embedCount++
				}
			})

			linkDensity := tr.getLinkDensity(node)
			contentLength := strLen(node.InnerText())
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
			} else if (embedCount == 1 && contentLength < 75) || embedCount > 1 {
				toRemove = true
			}
			if toRemove {
				node.Remove()
			}
		}
	})
}

// func (tr *TReadability) cleanStyle(e *goquery.Selection) {
// 	if e == nil {
// 		return
// 	}
// 	e.Find("*").Each(func(i int, elem *goquery.Selection) {
// 		// elem.RemoveAttr("class")
// 		// elem.RemoveAttr("id")
// 		// elem.RemoveAttr("style")
// 		// elem.RemoveAttr("width")
// 		// elem.RemoveAttr("height")
// 		// elem.RemoveAttr("onclick")
// 		// elem.RemoveAttr("onmouseover")
// 		// elem.RemoveAttr("border")
// 	})
// }

func (tr *TReadability) clean(node *Node, tag string) {
	if node == nil {
		return
	}
	// isEmbed := false
	// if tag == "embed" || tag == "object" {
	// 	isEmbed = true
	// }
	node.WorkElementNode(tag, func(n *Node) {
		n.Remove()
	})
}

func (tr *TReadability) cleanArticle(content *Node) {
	if content == nil {
		return
	}

	// tr.cleanStyle(content)
	tr.clean(content, "object")
	tr.cleanConditionally(content, "form")

	if content.TagLength("h1") == 1 {
		tr.clean(content, "h1")
	}
	if content.TagLength("h2") == 1 {
		tr.clean(content, "h2")
	}
	if content.TagLength("h3") == 1 {
		tr.clean(content, "h3")
	}

	tr.cleanConditionally(content, "table")
	tr.cleanConditionally(content, "ul")
	tr.cleanConditionally(content, "div")

	content.WorkElementNode("p", func(s *Node) {
		imgCount := s.TagLength("img")
		embedCount := s.TagLength("embed")
		objectCount := s.TagLength("object")
		if imgCount == 0 && embedCount == 0 && objectCount == 0 && s.InnerText() == "" {
			s.Remove()
		}
	})
	// tr.fixImagesPath(content)
	// tr.fixHrefPath(content)
	// tr.fixIframePath(content)

	// summary := ""
	// content.Find("*").Each(func(i int, s *goquery.Selection) {
	// 	if CheckHide(s) {
	// 		s.Find("*").Each(func(i int, z *goquery.Selection) {
	// 			s.Remove()
	// 		})
	// 		return
	// 	}
	// })

	// content.Find("p").Each(func(i int, s *goquery.Selection) {
	// 	summary = summary + s.Text() // TODO 很奇怪
	// })
	// tr.Summary = summary

	html, err := content.Html()
	if err != nil {
		return
	}
	// html = ghtml.UnescapeString(html)
	tr.Content = killBreaks.ReplaceAllString(html, "<br />")
	return
}
