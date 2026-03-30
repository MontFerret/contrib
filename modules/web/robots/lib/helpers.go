package lib

import (
	"github.com/MontFerret/contrib/modules/web/robots/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func decodeDocument(value runtime.Value) (core.Document, error) {
	var doc core.Document

	if err := sdk.Decode(value, &doc); err != nil {
		return core.Document{}, err
	}

	return doc, nil
}

func encodeValue(input any) runtime.Value {
	return sdk.Encode(input)
}
