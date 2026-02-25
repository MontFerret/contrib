package http

import (
	"context"
	"hash/fnv"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/PuerkitoBio/goquery"
)

type HTMLPage struct {
	document *HTMLDocument
	cookies  *drivers.HTTPCookies
	frames   *runtime.Array
	response drivers.HTTPResponse
}

func NewHTMLPage(
	qdoc *goquery.Document,
	url string,
	response drivers.HTTPResponse,
	cookies *drivers.HTTPCookies,
) (*HTMLPage, error) {
	doc, err := NewRootHTMLDocument(qdoc, url)

	if err != nil {
		return nil, err
	}

	p := new(HTMLPage)
	p.document = doc
	p.cookies = cookies
	p.frames = nil
	p.response = response

	return p, nil
}

func (p *HTMLPage) MarshalJSON() ([]byte, error) {
	return p.document.MarshalJSON()
}

func (p *HTMLPage) Type() runtime.Type {
	return drivers.HTMLPageType
}

func (p *HTMLPage) String() string {
	return p.document.GetURL().String()
}

func (p *HTMLPage) Compare(other runtime.Value) int64 {
	typed, ok := other.(runtime.Typed)

	if !ok {
		return 1
	}

	tc := drivers.Compare(p.Type(), typed.Type())

	if tc != 0 {
		return tc
	}

	httpPage, ok := other.(*HTMLPage)

	if !ok {
		return 1
	}

	return p.document.GetURL().Compare(httpPage.GetURL())
}

func (p *HTMLPage) Unwrap() any {
	return p.document
}

func (p *HTMLPage) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte("HTTP"))
	h.Write([]byte(p.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(p.document.GetURL()))

	return h.Sum64()
}

func (p *HTMLPage) Copy() runtime.Value {
	var cookies *drivers.HTTPCookies

	if p.cookies != nil {
		cookies = p.cookies.Copy().(*drivers.HTTPCookies)
	}

	page, err := NewHTMLPage(
		p.document.doc,
		p.document.GetURL().String(),
		p.response,
		cookies,
	)

	if err != nil {
		return runtime.None
	}

	return page
}

func (p *HTMLPage) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return p.document.Iterate(ctx)
}

func (p *HTMLPage) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return common.GetInPage(ctx, key, p)
}

func (p *HTMLPage) Set(ctx context.Context, key, value runtime.Value) error {
	return common.SetInPage(ctx, key, p, value)
}

func (p *HTMLPage) Length(ctx context.Context) (runtime.Int, error) {
	return p.document.Length(ctx)
}

func (p *HTMLPage) Close() error {
	return nil
}

func (p *HTMLPage) IsClosed() runtime.Boolean {
	return runtime.True
}

func (p *HTMLPage) GetURL() runtime.String {
	return p.document.GetURL()
}

func (p *HTMLPage) GetMainFrame() drivers.HTMLDocument {
	return p.document
}

func (p *HTMLPage) GetFrames(ctx context.Context) (runtime.List, error) {
	if p.frames == nil {
		arr := runtime.NewArray(10)

		err := common.CollectFrames(ctx, arr, p.document)

		if err != nil {
			return runtime.NewArray(0), err
		}

		p.frames = arr
	}

	return p.frames, nil
}

func (p *HTMLPage) GetFrame(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	if p.frames == nil {
		arr := runtime.NewArray(10)

		err := common.CollectFrames(ctx, arr, p.document)

		if err != nil {
			return runtime.None, err
		}

		p.frames = arr
	}

	return p.frames.Get(ctx, idx)
}

func (p *HTMLPage) GetCookies(_ context.Context) (*drivers.HTTPCookies, error) {
	res := drivers.NewHTTPCookies()

	if p.cookies != nil {
		p.cookies.ForEach(func(value drivers.HTTPCookie, _ runtime.String) bool {
			res.Set(value)

			return true
		})
	}

	return res, nil
}

func (p *HTMLPage) GetResponse(_ context.Context) (drivers.HTTPResponse, error) {
	return p.response, nil
}

func (p *HTMLPage) SetCookies(_ context.Context, _ *drivers.HTTPCookies) error {
	return runtime.ErrNotSupported
}

func (p *HTMLPage) DeleteCookies(_ context.Context, _ *drivers.HTTPCookies) error {
	return runtime.ErrNotSupported
}

func (p *HTMLPage) PrintToPDF(_ context.Context, _ drivers.PDFParams) (runtime.Binary, error) {
	return nil, runtime.ErrNotSupported
}

func (p *HTMLPage) CaptureScreenshot(_ context.Context, _ drivers.ScreenshotParams) (runtime.Binary, error) {
	return nil, runtime.ErrNotSupported
}

func (p *HTMLPage) WaitForNavigation(_ context.Context, _ runtime.String) error {
	return runtime.ErrNotSupported
}

func (p *HTMLPage) WaitForFrameNavigation(_ context.Context, _ drivers.HTMLDocument, _ runtime.String) error {
	return runtime.ErrNotSupported
}

func (p *HTMLPage) Navigate(_ context.Context, _ runtime.String) error {
	return runtime.ErrNotSupported
}

func (p *HTMLPage) NavigateBack(_ context.Context, _ runtime.Int) (runtime.Boolean, error) {
	return false, runtime.ErrNotSupported
}

func (p *HTMLPage) NavigateForward(_ context.Context, _ runtime.Int) (runtime.Boolean, error) {
	return false, runtime.ErrNotSupported
}

func (p *HTMLPage) Subscribe(_ context.Context, _ runtime.Subscription) (runtime.Stream, error) {
	return nil, runtime.ErrNotSupported
}

func (p *HTMLPage) Dispatch(ctx context.Context, event runtime.DispatchEvent) (runtime.Value, error) {
	//TODO implement me
	panic("implement me")
}
