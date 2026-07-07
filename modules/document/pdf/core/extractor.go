package core

import (
	"fmt"

	ledongpdf "github.com/ledongthuc/pdf"
)

func pageContent(page ledongpdf.Page) (content ledongpdf.Content, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("extract page content: %v", r)
		}
	}()

	return page.Content(), nil
}
