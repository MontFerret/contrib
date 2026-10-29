package main

import (
	"encoding/json"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type Params []string

func (p *Params) String() string {
	return "[" + strings.Join(*p, ",") + "]"
}

func (p *Params) Set(value string) error {
	*p = append(*p, value)

	return nil
}

func (p *Params) ToMap() (runtime.Params, error) {
	res := runtime.NewParams()

	for _, entry := range *p {
		pair := strings.SplitN(entry, ":", 2)

		if len(pair) < 2 {
			return nil, runtime.Error(runtime.ErrInvalidArgument, entry)
		}

		var value interface{}
		key := pair[0]

		err := json.Unmarshal([]byte(pair[1]), &value)
		if err != nil {
			return nil, err
		}

		if err := res.Set(key, value); err != nil {
			return nil, err
		}
	}

	return res, nil
}
