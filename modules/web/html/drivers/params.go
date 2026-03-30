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
		Cookies     *HTTPCookies `json:"cookies"`
		Headers     *HTTPHeaders `json:"headers"`
		Viewport    *Viewport    `json:"viewport"`
		Ignore      *Ignore      `json:"ignore"`
		URL         string       `json:"url"`
		UserAgent   string       `json:"userAgent"`
		Charset     string       `json:"charset"`
		KeepCookies bool         `json:"keepCookies"`
	}

	ParseParams struct {
		Cookies     *HTTPCookies `json:"cookies"`
		Headers     *HTTPHeaders `json:"headers"`
		Viewport    *Viewport    `json:"viewport"`
		Content     []byte       `json:"content"`
		KeepCookies bool         `json:"keepCookies"`
	}
)
