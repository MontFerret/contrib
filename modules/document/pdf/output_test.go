package pdf

import (
	"context"
	"errors"

	"github.com/MontFerret/ferret/v2"
)

// collectResult observes terminal execution errors even for side-effect-only queries.
func collectResult(output ferret.Output, err error) (content *ferret.Content, resultErr error) {
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := output.Close(); closeErr != nil {
			if resultErr == nil {
				resultErr = closeErr
			} else {
				resultErr = errors.Join(resultErr, closeErr)
			}
		}
	}()

	return output.Collect(context.Background())
}
