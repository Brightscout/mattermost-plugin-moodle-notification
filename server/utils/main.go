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
