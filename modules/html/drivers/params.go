package drivers

type (
	ResourceFilter struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	StatusCodeFilter struct {
		URL  string `json:"url"`
		Code int    `json:"code"`
	}

	Ignore struct {
		Resources   []ResourceFilter   `json:"resources"`
		StatusCodes []StatusCodeFilter `json:"statusCodes"`
	}

	Viewport struct {
		Height      int     `json:"height"`
		Width       int     `json:"width"`
		ScaleFactor float64 `json:"scaleFactor"`
		Mobile      bool    `json:"mobile"`
		Landscape   bool    `json:"landscape"`
	}

	Params struct {
		URL         string       `json:"url"`
		UserAgent   string       `json:"userAgent"`
		KeepCookies bool         `json:"keepCookies"`
		Cookies     *HTTPCookies `json:"cookies"`
		Headers     *HTTPHeaders `json:"headers"`
		Viewport    *Viewport    `json:"viewport"`
		Charset     string       `json:"charset"`
		Ignore      *Ignore      `json:"ignore"`
	}

	ParseParams struct {
		Content     []byte       `json:"content"`
		KeepCookies bool         `json:"keepCookies"`
		Cookies     *HTTPCookies `json:"cookies"`
		Headers     *HTTPHeaders `json:"headers"`
		Viewport    *Viewport    `json:"viewport"`
	}
)
