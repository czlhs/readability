package readability

import (
	"math"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (self *TReadability) extract() *goquery.Selection {
	self.htmlDoc.Find("p").Each(func(i int, node *goquery.Selection) {
		parentNode := node.Parent()
		grandParentNode := parentNode.Parent()
		innerText := node.Text()

		if parentNode == nil || strLen(innerText) < 20 {
			return
		}
		parentHash := HashStr(parentNode)
		grandParentHash := HashStr(grandParentNode)
		if _, ok := self.candidates[parentHash]; !ok {
			self.candidates[parentHash] = self.initializeNode(parentNode)
		}
		if _, ok := self.candidates[grandParentHash]; !ok {
			self.candidates[grandParentHash] = self.initializeNode(grandParentNode)
		}
		contentScore := 1.0
		contentScore += float64(strings.Count(innerText, ","))
		contentScore += float64(strings.Count(innerText, "，"))
		contentScore += math.Min(math.Floor(float64(strLen(innerText)/100)), 3)

		v, _ := self.candidates[parentHash]
		v.score += contentScore
		self.candidates[parentHash] = v

		if grandParentNode != nil {
			v, _ = self.candidates[grandParentHash]
			v.score += contentScore / 2.0
			self.candidates[grandParentHash] = v
		}
	})

	var topCandidate *TCandidateItem
	for k, v := range self.candidates {
		v.score = v.score * (1 - self.getLinkDensity(v.node))
		self.candidates[k] = v

		if topCandidate == nil || v.score > topCandidate.score {
			if topCandidate == nil {
				topCandidate = new(TCandidateItem)
			}
			topCandidate.score = v.score
			topCandidate.node = v.node
		}
	}
	if topCandidate != nil {
		//		fmt.Println("topCandidate.score=", topCandidate.score)
		return topCandidate.node
		// return self.cleanArticle(topCandidate.node)
	}
	return nil
}

func (self *TReadability) initializeNode(node *goquery.Selection) TCandidateItem {
	contentScore := 0.0
	switch self.getTagName(node) {
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
	contentScore += self.getClassWeight(node)
	return TCandidateItem{contentScore, node}
}

func (self *TReadability) getClassWeight(node *goquery.Selection) float64 {
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

// 链接减分
func (self *TReadability) getLinkDensity(node *goquery.Selection) float64 {
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
			linkLength += float64(strLen(link.Text()))
		})
	return linkLength / textLength
}
