package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestEncodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("encodes representative object and round trips", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"title": runtime.NewString("Ferret"),
			"server": runtime.NewObjectWith(map[string]runtime.Value{
				"host": runtime.NewString("localhost"),
				"port": runtime.NewInt(8080),
			}),
			"plugins": runtime.NewArrayWith(
				runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("html")}),
				runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("json")}),
			),
		})

		result, err := Encode(ctx, value, DefaultEncodeOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		decoded, err := Decode(ctx, runtime.NewString(result), DefaultDecodeOptions())
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if decoded.String() != value.String() {
			t.Fatalf("round-trip mismatch: got %s want %s", decoded.String(), value.String())
		}
	})

	t.Run("sorts keys when requested", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"b": runtime.NewInt(2),
			"a": runtime.NewInt(1),
		})

		opts := DefaultEncodeOptions()
		opts.SortKeys = true

		result, err := Encode(ctx, value, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(result, "a = 1\nb = 2") {
			t.Fatalf("expected sorted output, got %q", result)
		}
	})

	t.Run("encodes arrays of tables", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"plugins": runtime.NewArrayWith(
				runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("html")}),
				runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("json")}),
			),
		})

		result, err := Encode(ctx, value, DefaultEncodeOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "[[plugins]]") {
			t.Fatalf("expected array-of-tables header, got %q", result)
		}
	})

	t.Run("preserves local temporal forms", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"local_dt":   runtime.NewDateTime(time.Date(1979, 5, 27, 7, 32, 0, 0, time.FixedZone(localDateTimeLocation, -4*60*60))),
			"local_date": runtime.NewDateTime(time.Date(1979, 5, 27, 0, 0, 0, 0, time.FixedZone(localDateLocation, -4*60*60))),
			"local_time": runtime.NewDateTime(time.Date(0, 1, 1, 7, 32, 0, 0, time.FixedZone(localTimeLocation, -4*60*60))),
		})

		opts := DefaultEncodeOptions()
		opts.DateTime = EncodeDateTimePreserve
		opts.SortKeys = true

		result, err := Encode(ctx, value, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "local_dt = 1979-05-27T07:32:00") {
			t.Fatalf("expected preserved local datetime, got %q", result)
		}

		if !strings.Contains(result, "local_date = 1979-05-27") {
			t.Fatalf("expected preserved local date, got %q", result)
		}

		if !strings.Contains(result, "local_time = 07:32:00") {
			t.Fatalf("expected preserved local time, got %q", result)
		}
	})

	t.Run("encodes native datetimes as rfc3339", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"released": runtime.NewDateTime(time.Date(1979, 5, 27, 7, 32, 0, 0, time.FixedZone(localDateTimeLocation, -4*60*60))),
		})

		result, err := Encode(ctx, value, DefaultEncodeOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "released = 1979-05-27T07:32:00-04:00") {
			t.Fatalf("expected RFC3339 datetime, got %q", result)
		}
	})

	t.Run("rejects top-level non object", func(t *testing.T) {
		_, err := Encode(ctx, runtime.NewArrayWith(runtime.NewInt(1)), DefaultEncodeOptions())
		if err == nil {
			t.Fatal("expected top-level object error")
		}

		if _, ok := err.(*Error); !ok {
			t.Fatalf("expected *TOMLError, got %T", err)
		}
	})

	t.Run("rejects none values", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"maybe": runtime.None,
		})

		_, err := Encode(ctx, value, DefaultEncodeOptions())
		if err == nil {
			t.Fatal("expected none encoding error")
		}
	})

	t.Run("rejects mixed arrays", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"items": runtime.NewArrayWith(
				runtime.NewInt(1),
				runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("bad")}),
			),
		})

		_, err := Encode(ctx, value, DefaultEncodeOptions())
		if err == nil {
			t.Fatal("expected mixed array error")
		}
	})

	t.Run("rejects nested arrays with objects", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"items": runtime.NewArrayWith(
				runtime.NewArrayWith(
					runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.NewString("bad")}),
				),
			),
		})

		_, err := Encode(ctx, value, DefaultEncodeOptions())
		if err == nil {
			t.Fatal("expected nested object array error")
		}
	})

	t.Run("rejects binary values without panicking", func(t *testing.T) {
		value := runtime.NewObjectWith(map[string]runtime.Value{
			"blob": runtime.NewBinary([]byte("toml")),
		})

		_, err := Encode(ctx, value, DefaultEncodeOptions())
		if err == nil {
			t.Fatal("expected binary encoding error")
		}

		if !strings.Contains(err.Error(), "Binary") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
