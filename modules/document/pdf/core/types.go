package core

// PageInfo describes normalized page metadata exposed to Ferret.
type PageInfo struct {
	Width    float64
	Height   float64
	Number   int
	Rotation int
}

// Bounds describes a text fragment box in PDF points.
type Bounds struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// TextBlock is a low-level positioned text fragment.
type TextBlock struct {
	Text   string
	Bounds Bounds
}
