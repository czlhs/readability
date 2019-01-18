// 改自 https://github.com/kingwkb/readability python版本
// 于2016-11-10
// by: ying32
package readability

import (
	"regexp"
)

var (
	unlikelyCandidates   = regexp.MustCompile(`(?is)combx|comment|community|disqus|extra|foot|header|menu|remark|rss|shoutbox|sidebar|sponsor|ad-break|agegate|pagination|pager|popup|tweet|twitter|location`)
	okMaybeItsACandidate = regexp.MustCompile(`(?ims)and|article|body|column|main|shadow|story|entry|^post`)
	positive             = regexp.MustCompile(`(?is)article|body|content|entry|hentry|main|page|pagination|post|text|blog|story`)
	negative             = regexp.MustCompile(`(?is)combx|comment|com|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tags|tool|widget`)
	extraneous           = regexp.MustCompile(`(?is)print|archive|comment|discuss|e[\-]?mail|share|reply|all|login|sign|single`)
	divToPElements       = regexp.MustCompile(`(?is)<(a|blockquote|dl|div|img|ol|p|pre|table|ul)`)
	replaceBrs           = regexp.MustCompile(`(?is)(<br[^>]*>[ \n\r\t]*){2,}`)
	replaceFonts         = regexp.MustCompile(`(?is)<(/?)font[^>]*>`)
	trim                 = regexp.MustCompile(`(?is)^\s+|\s+$`)
	normalize            = regexp.MustCompile(`(?is)\s{2,}`)
	killBreaks           = regexp.MustCompile(`(?is)(<br\s*/?>(\s|&nbsp;?)*)+`)
	videos               = regexp.MustCompile(`(?is)http://(www\.)?(youtube|vimeo)\.com`)
	attributeScore       = regexp.MustCompile(`(?is)blog|post|article`)
	skipFootnoteLink     = regexp.MustCompile(`(?is)^\s*(\[?[a-z0-9]{1,2}\]?|^|edit|citation needed)\s*$"`)
	nextLink             = regexp.MustCompile(`(?is)(next|weiter|continue|>([^\|]|$)|»([^\|]|$))`)
	prevLink             = regexp.MustCompile(`(?is)(prev|earl|old|new|<|«)`)

	unlikelyElements = regexp.MustCompile(`(?is)(input|time|button)`)

	contentMayContain = regexp.MustCompile(`\.( |$)`)
	pageCodeReg       = regexp.MustCompile(`(?is)<meta.+?charset=[^\w]?([-\w]+)`)
	validURL          = regexp.MustCompile(`^(https?)?://(www\.)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
)
