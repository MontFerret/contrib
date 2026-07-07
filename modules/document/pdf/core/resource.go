package core

import commonresource "github.com/MontFerret/contrib/pkg/common/resource"

var resourceIDs commonresource.IDGenerator

func newResourceID() uint64 {
	return resourceIDs.Next()
}
