package core

import (
	"errors"
	"math"

	ledongpdf "github.com/ledongthuc/pdf"
)

func pageInfo(page ledongpdf.Page, number int) (PageInfo, error) {
	box, ok := pageBox(page, "CropBox")
	if !ok {
		box, ok = pageBox(page, "MediaBox")
	}
	if !ok {
		return PageInfo{}, errors.New("page box is missing or invalid")
	}

	return PageInfo{
		Number:   number,
		Width:    math.Abs(box[2] - box[0]),
		Height:   math.Abs(box[3] - box[1]),
		Rotation: pageRotation(page),
	}, nil
}

func pageBox(page ledongpdf.Page, name string) ([4]float64, bool) {
	value := findInherited(page.V, name)
	if value.IsNull() || value.Len() < 4 {
		return [4]float64{}, false
	}

	box := [4]float64{
		value.Index(0).Float64(),
		value.Index(1).Float64(),
		value.Index(2).Float64(),
		value.Index(3).Float64(),
	}

	return box, math.Abs(box[2]-box[0]) > 0 && math.Abs(box[3]-box[1]) > 0
}

func pageRotation(page ledongpdf.Page) int {
	value := findInherited(page.V, "Rotate")
	if value.IsNull() {
		return 0
	}

	rotation := int(value.Int64() % 360)
	if rotation < 0 {
		rotation += 360
	}

	return rotation
}

func findInherited(value ledongpdf.Value, key string) ledongpdf.Value {
	for current := value; !current.IsNull(); current = current.Key("Parent") {
		found := current.Key(key)
		if !found.IsNull() {
			return found
		}
	}

	return ledongpdf.Value{}
}
