package html

//import (
//	"context"
//	"testing"
//
//	. "github.com/smartystreets/goconvey/convey"
//
//	"github.com/MontFerret/ferret/v2/pkg/runtime"
//	"github.com/MontFerret/ferret/v2/pkg/stdlib/html"
//)
//
//func TestDocumentExists(t *testing.T) {
//	Convey("DOCUMENT_EXISTS", t, func() {
//		Convey("Should return error if url is not a string", func() {
//			_, err := html.DocumentExists(context.Background(), runtime.None)
//
//			So(err, ShouldNotBeNil)
//		})
//
//		Convey("Should return error if options is not an object", func() {
//			_, err := html.DocumentExists(context.Background(), runtime.NewString("http://fsdfsdfdsdsf.fdf"), runtime.None)
//
//			So(err, ShouldNotBeNil)
//		})
//
//		Convey("Should return error if headers is not an object", func() {
//			opts := runtime.NewObjectWith(runtime.NewObjectProperty("headers", runtime.None))
//			_, err := html.DocumentExists(context.Background(), runtime.NewString("http://fsdfsdfdsdsf.fdf"), opts)
//
//			So(err, ShouldNotBeNil)
//		})
//
//		Convey("Should return 'false' when a website does not exist by a given url", func() {
//			out, err := html.DocumentExists(context.Background(), runtime.NewString("http://fsdfsdfdsdsf.fdf"))
//
//			So(err, ShouldBeNil)
//			So(out, ShouldEqual, runtime.False)
//		})
//
//		Convey("Should return 'true' when a website exists by a given url", func() {
//			out, err := html.DocumentExists(context.Background(), runtime.NewString("https://www.google.com/"))
//
//			So(err, ShouldBeNil)
//			So(out, ShouldEqual, runtime.True)
//		})
//	})
//}
