package core

import "fmt"

func validateSheetName(name string) error {
	if name == "" {
		return fmt.Errorf("worksheet name must not be empty")
	}

	return nil
}
