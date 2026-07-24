package llm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	module := New()
	if module == nil {
		t.Fatal("expected module to be non-nil")
	}
	if module.Name() != "ai/llm" {
		t.Fatalf("expected module name %q, got %q", "ai/llm", module.Name())
	}
}

func TestModuleAppliesProviderOverridesInOptionOrder(t *testing.T) {
	engine, err := ferret.New(ferret.WithModules(New(
		WithProviderFactory(newFakeProviderFactoryWithGeneration("first factory")),
		WithProviderFactory(newFakeProviderFactoryWithGeneration("last factory")),
	)))
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	output, err := engine.Run(context.Background(), source.NewAnonymous(`
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		RETURN AI::LLM::GENERATE(model, "hello")
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual string
	decodeOutput(t, output.Content, &actual)
	if actual != "last factory" {
		t.Fatalf("expected last same-name provider override, got %q", actual)
	}
}

func TestModuleInstallsConfiguredRegistryInRunContext(t *testing.T) {
	var observed *core.Registry
	callerRegistry := core.NewRegistry()
	if err := callerRegistry.Register(newFakeProviderFactoryWithGeneration("caller factory")); err != nil {
		t.Fatalf("unexpected caller provider registration error: %v", err)
	}

	engine, err := ferret.New(
		ferret.WithAfterRunHook(func(ctx context.Context, _ error) error {
			var lookupErr error
			observed, lookupErr = core.RegistryFrom(ctx)

			return lookupErr
		}),
		ferret.WithModules(New(
			WithProviderFactory(newFakeProviderFactoryWithGeneration("configured factory")),
		)),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	runCtx := core.WithRegistry(context.Background(), callerRegistry)
	if _, err := engine.Run(runCtx, source.NewAnonymous("RETURN true")); err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if observed == nil {
		t.Fatal("expected after-run hook to observe the provider registry")
	}

	model, err := observed.NewModel(
		context.Background(),
		"openai",
		core.ModelOptions{Model: "opaque/model-name", APIKey: "explicit-test-key"},
	)
	if err != nil {
		t.Fatalf("unexpected model creation error: %v", err)
	}

	response, err := model.Generate(context.Background(), core.Request{})
	if err != nil {
		t.Fatalf("unexpected generation error: %v", err)
	}
	if response.Text != "configured factory" {
		t.Fatalf("expected configured provider factory, got %q", response.Text)
	}
}

func TestModuleRunsGenerationFunction(t *testing.T) {
	output := runFQL(t, `
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		RETURN AI::LLM::GENERATE(model, "hello")
	`)

	var actual string
	decodeOutput(t, output.Content, &actual)
	if actual != "generated response" {
		t.Fatalf("unexpected generation result: %q", actual)
	}
}

func TestModuleRunsPlainAndScalarQueries(t *testing.T) {
	output := runFQL(t, `
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		LET many = QUERY "query list" IN model
		LET one = QUERY ONE "query scalar" IN model USING generate
		RETURN {many: many, one: one}
	`)

	var actual struct {
		One  string   `json:"one"`
		Many []string `json:"many"`
	}
	decodeOutput(t, output.Content, &actual)

	if len(actual.Many) != 1 || actual.Many[0] != "generated response" {
		t.Fatalf("unexpected plain QUERY result: %#v", actual.Many)
	}
	if actual.One != "scalar answer" {
		t.Fatalf("unexpected QUERY ONE result: %q", actual.One)
	}
}

func TestModuleRunsTwoTurnLocalSession(t *testing.T) {
	output := runFQL(t, `
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		LET session = AI::LLM::SESSION(model, {
			instructions: "Be concise.",
			context: {
				mode: "local",
				overflow: "error",
				maxTokens: 128000,
				reserveOutputTokens: 4000
			}
		})
		LET first = AI::LLM::CHAT(session, "first turn")
		LET second = AI::LLM::CHAT(session, "second turn")
		RETURN {
			first: first,
			second: second,
			history: AI::LLM::HISTORY(session)
		}
	`)

	var actual struct {
		First   string `json:"first"`
		Second  string `json:"second"`
		History []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"history"`
	}
	decodeOutput(t, output.Content, &actual)

	if actual.First != "first answer" || actual.Second != "second answer" {
		t.Fatalf("unexpected session responses: %#v", actual)
	}
	if len(actual.History) != 4 {
		t.Fatalf("expected four committed messages, got %#v", actual.History)
	}

	want := []struct {
		role    string
		content string
	}{
		{role: "user", content: "first turn"},
		{role: "assistant", content: "first answer"},
		{role: "user", content: "second turn"},
		{role: "assistant", content: "second answer"},
	}
	for index, expected := range want {
		message := actual.History[index]
		if message.Role != expected.role || message.Content != expected.content {
			t.Fatalf("unexpected history[%d]: %#v", index, message)
		}
	}
}

func TestModuleRunsSummarizeAndReset(t *testing.T) {
	factory := newFakeProviderFactory()
	engine, err := ferret.New(ferret.WithModules(New(WithProviderFactory(factory))))
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	output, err := engine.Run(context.Background(), source.NewAnonymous(`
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		LET summary = AI::LLM::SUMMARIZE(model, "long source text", {
			style: "concise",
			maxWords: 25,
			instructions: "Prefer active voice.",
			temperature: 0,
			maxOutputTokens: 100,
			timeout: 0
		})
		LET session = AI::LLM::SESSION(model, {})
		LET ignored = AI::LLM::CHAT(session, "first turn")
		LET reset = AI::LLM::RESET(session)
		RETURN {
			summary: summary,
			reset: reset,
			history: AI::LLM::HISTORY(session)
		}
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var actual struct {
		Summary string `json:"summary"`
		History []any  `json:"history"`
		Reset   bool   `json:"reset"`
	}
	decodeOutput(t, output.Content, &actual)
	if actual.Summary != "generated response" || !actual.Reset || len(actual.History) != 0 {
		t.Fatalf("unexpected summarize/reset result: %#v", actual)
	}

	factory.backend.mu.Lock()
	if len(factory.backend.requests) < 1 {
		factory.backend.mu.Unlock()
		t.Fatal("expected summarized provider request")
	}
	request := factory.backend.requests[0]
	factory.backend.mu.Unlock()

	if request.Options.Temperature == nil || *request.Options.Temperature != 0 {
		t.Fatalf("expected explicit zero temperature, got %#v", request.Options.Temperature)
	}
	if request.Options.MaxOutputTokens != 100 {
		t.Fatalf("unexpected output token limit: %d", request.Options.MaxOutputTokens)
	}
	if request.Instructions == "" {
		t.Fatal("expected deterministic summarize instructions")
	}
}

func TestModuleRunsExtractionAndClassification(t *testing.T) {
	output := runFQL(t, `
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		LET extracted = AI::LLM::EXTRACT(model, "Ada scored nine", {
			type: "object",
			properties: {
				name: {type: "string"},
				score: {type: "integer"}
			},
			required: ["name", "score"],
			additionalProperties: false
		})
		LET classified = AI::LLM::CLASSIFY(
			model,
			"I was charged twice",
			["billing", "technical"]
		)
		RETURN {extracted: extracted, classified: classified}
	`)

	var actual struct {
		Classified struct {
			Label string `json:"label"`
		} `json:"classified"`
		Extracted struct {
			Name  string `json:"name"`
			Score int    `json:"score"`
		} `json:"extracted"`
	}
	decodeOutput(t, output.Content, &actual)

	if actual.Extracted.Name != "Ada" || actual.Extracted.Score != 9 {
		t.Fatalf("unexpected extraction result: %#v", actual.Extracted)
	}
	if actual.Classified.Label != "billing" {
		t.Fatalf("unexpected classification result: %#v", actual.Classified)
	}
}

func TestModuleClosesEveryExecutionScopedSession(t *testing.T) {
	var observed *core.SessionScope
	engine, err := ferret.New(
		ferret.WithAfterRunHook(func(ctx context.Context, _ error) error {
			observed, _ = core.SessionScopeFrom(ctx)

			return nil
		}),
		ferret.WithModules(New(WithProviderFactory(newFakeProviderFactory()))),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	_, err = engine.Run(context.Background(), source.NewAnonymous(`
		LET shortcut = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key",
			session: true
		})
		LET model = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key"
		})
		LET session = AI::LLM::SESSION(model, {})
		LET fork = AI::LLM::FORK(session)
		RETURN {
			shortcut: [shortcut],
			nested: {session: session, fork: fork}
		}
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if observed == nil {
		t.Fatal("expected after-run hook to observe the session scope")
	}
	if observed.Len() != 0 {
		t.Fatalf("expected session scope cleanup, got %d tracked sessions", observed.Len())
	}

	observed = nil
	_, err = engine.Run(context.Background(), source.NewAnonymous(`
		LET session = AI::LLM::MODEL("openai", {
			model: "opaque/model-name",
			apiKey: "explicit-test-key",
			session: true
		})
		RETURN QUERY COUNT "unsupported" IN session
	`))
	if err == nil {
		t.Fatal("expected unsupported QUERY COUNT error")
	}
	if observed == nil || observed.Len() != 0 {
		t.Fatalf("expected failed-run scope cleanup, got %#v", observed)
	}
}

func runFQL(t *testing.T, query string) *ferret.Output {
	t.Helper()

	engine, err := ferret.New(ferret.WithModules(New(WithProviderFactory(newFakeProviderFactory()))))
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	output, err := engine.Run(context.Background(), source.NewAnonymous(query))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	return output
}

func decodeOutput(t *testing.T, data []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("failed to decode output: %v", err)
	}
}
