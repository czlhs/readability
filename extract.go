package readability

import (
	"strings"
)

func (self *TReadability) initializeNode(node *Node) TCandidateItem {
	contentScore := 0.0
	switch node.GetTagName() {
	case "article":
		contentScore += 10
	case "section":
		contentScore += 8
	case "div":
		contentScore += 5
	case "pre", "blockquote", "td":
		contentScore += 3
	case "form", "ul", "ol", "dl", "dd", "dt", "li", "address":
		contentScore -= 3
	case "th", "h1", "h2", "h3", "h4", "h5", "h6":
		contentScore -= 5
	}
	// TODO node.attributes
	contentScore += self.getClassWeight(node)
	return TCandidateItem{contentScore, node}
}

func (self *TReadability) getClassWeight(node *Node) float64 {
	weight := 0.0
	if str := node.GetAttr("class"); str != "" {
		if negative.MatchString(str) {
			weight -= 25
		}
		if positive.MatchString(str) {
			weight += 25
		}
	}
	if str := node.GetAttr("id"); str != "" {
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
func (self *TReadability) getLinkDensity(node *Node) float64 {
	if node == nil {
		return 0
	}
	textLength := float64(strLen(node.InnerText()))
	if textLength == 0 {
		return 0
	}
	linkLength := 0.0
	node.WorkElementNode("a", func(n *Node) {
		if href := n.GetAttr("href"); href != "" && !strings.HasPrefix(href, "#") {
			linkLength += float64(strLen(n.InnerText()))
		}

	})
	return linkLength / textLength
}
