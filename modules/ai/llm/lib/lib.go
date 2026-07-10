package lib

import (
	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// RegisterLib registers the AI::LLM namespace functions.
func RegisterLib(ns runtime.Namespace, registry *core.Registry) {
	ns.Function().A2().
		Add("MODEL", modelWithRegistry(registry)).
		Add("SESSION", Session)

	ns.Function().Var().
		Add("GENERATE", Generate).
		Add("CHAT", Chat).
		Add("SUMMARIZE", Summarize).
		Add("EXTRACT", Extract).
		Add("CLASSIFY", Classify)

	ns.Function().A1().
		Add("RESET", Reset).
		Add("FORK", Fork).
		Add("HISTORY", History)
}
