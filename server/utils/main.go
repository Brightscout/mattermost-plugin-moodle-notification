package utils

import (
	"fmt"
	"regexp"
)

// This function replaces all the embedded images in markdown with markdown link.
func ReplaceEmbeddedImages(md string) string {
	re := regexp.MustCompile(`!\[(?P<text>.*?)\]\((?P<protocol>\w+):(?P<host>[^\n)]+)?\)`)
	s := re.ReplaceAllStringFunc(md, func(str string) string {
		imgText := []byte(str)
		loc := re.FindAllSubmatchIndex(imgText, -1)[0]

		// if alt text is empty replace it with "view image"
		text := string(imgText[loc[2]:loc[3]])
		if text == "" {
			text = "view image"
		}
		protocol := string(imgText[loc[4]:loc[5]])
		host := string(imgText[loc[6]:loc[7]])

		// return markdown link
		return fmt.Sprintf(`[\[%s\]](%s:%s)`, text, protocol, host)
	})
	return s
}

// This function wrap <span> tag around html entity like &raquo; &nbsp;
func NormaliseHTMLEntities(text string) string {
	re := regexp.MustCompile(`&([a-z0-9]+|#[0-9]{1,6}|#x[0-9a-fA-F]{1,6});`)
	return re.ReplaceAllStringFunc(text, func(str string) string {
		return fmt.Sprintf("<span> %s </span>", str)
	})
}
