package utils

import "fmt"

func Hyperlink(url, text string) string {
	if text == "" {
		text = url
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}
