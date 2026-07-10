# AI::LLM

`AI::LLM` exposes text generation, chat, summarization, structured extraction,
and classification to Ferret queries. Version 1 supports OpenAI through the
Responses API and keeps all conversation state local to the current Ferret run.

## Registration

Register the module when constructing the Ferret engine:

```go
engine, err := ferret.New(
	ferret.WithModules(llm.New()),
)
```

The module name is `ai/llm`, and its functions are registered under the
`AI::LLM` namespace.

Embedders can install another provider implementation with
`llm.WithProviderFactory(factory)`. Provider registration is case-insensitive;
when a custom factory has the same normalized name as a built-in provider, the
custom factory deliberately replaces that built-in. This is intended for
controlled embedding and deterministic tests, not for passing arbitrary SDK
options through FQL.

## Credentials and models

Create an OpenAI model with an explicit model name and API key:

```fql
LET model = AI::LLM::MODEL("openai", {
  model: "gpt-5-mini",
  apiKey: @openaiApiKey
})
RETURN AI::LLM::GENERATE(model, "Write a one-sentence greeting.")
```

`model` and `apiKey` must be nonblank strings. Provider identifiers are
case-insensitive, but model names are passed to the provider unchanged. The
module never uses credentials from process environment variables and does not
expose credentials through value formatting, debugging, or JSON serialization.

Set `session: true` to return a default local session directly:

```fql
LET chat = AI::LLM::MODEL("openai", {
  model: "gpt-5-mini",
  apiKey: @openaiApiKey,
  session: true
})
RETURN AI::LLM::CHAT(chat, "Remember that my project is called Ferret.")
```

## Functions

| Function | Signature | Result |
|---|---|---|
| `MODEL` | `MODEL(provider, options)` | Stateless model, or a default session when `session: true` |
| `SESSION` | `SESSION(model, options)` | Local session |
| `GENERATE` | `GENERATE(target, prompt[, options])` | String |
| `CHAT` | `CHAT(target, message[, options])` | String |
| `SUMMARIZE` | `SUMMARIZE(target, text[, options])` | String |
| `EXTRACT` | `EXTRACT(target, text, schema[, options])` | Schema-matching Ferret value |
| `CLASSIFY` | `CLASSIFY(target, text, labels[, options])` | `{label: "..."}` |
| `RESET` | `RESET(session)` | `true` |
| `FORK` | `FORK(session)` | Independent session copied from the current history |
| `HISTORY` | `HISTORY(session)` | Copied array of `{role, content}` objects |

All unknown option keys are rejected. Common execution options for the five
generation functions are:

```fql
{
  temperature: 0.0,       // number in [0, 2]
  maxOutputTokens: 1000,  // positive integer
  timeout: 30000          // nonnegative milliseconds
}
```

An explicit zero temperature is sent to the provider rather than treated as an
omitted value.

Operation-specific options are:

- `GENERATE`: optional `instructions`.
- `CHAT`: optional `messages` and `instructions`. `messages` is an array of
  `{role, content}` objects; content must be a string and role must be
  `system`, `developer`, `user`, or `assistant`. The function's message string
  is appended as the final user message.
- `SUMMARIZE`: optional `style`, positive `maxWords`, and `instructions`.
- `EXTRACT`: optional `instructions`; its required schema is the third
  positional argument.
- `CLASSIFY`: optional `instructions`; its required labels are unique,
  nonempty strings in the third positional argument.

For example:

```fql
LET summary = AI::LLM::SUMMARIZE(model, article, {
  style: "bullet points",
  maxWords: 120,
  temperature: 0,
  maxOutputTokens: 600
})

LET product = AI::LLM::EXTRACT(model, description, {
  type: "object",
  properties: {
    name: {type: "string"},
    price: {type: "number"}
  },
  required: ["name", "price"],
  additionalProperties: false
})

LET category = AI::LLM::CLASSIFY(
  model,
  ticket,
  ["billing", "technical", "account"]
)
```

Extraction schemas are compiled before a provider request. Local fragment
references such as `#/$defs/address` are allowed, while external `$ref` values
are rejected. Provider output is parsed and validated locally before it is
returned or added to session history.

## Query API

Models and sessions are queryable. An empty `USING` clause and
`USING generate` both perform generation:

```fql
LET model = AI::LLM::MODEL("openai", {
  model: "gpt-5-mini",
  apiKey: @openaiApiKey
})

LET generated = QUERY ONE "Write a haiku about parsers." IN model
LET summary = QUERY ONE article IN model USING summarize
  WITH {style: "concise", maxWords: 80}
  OPTIONS {temperature: 0, maxOutputTokens: 300, timeout: 30000}

RETURN {generated: generated, summary: summary}
```

`WITH` contains the operation-specific data described above; `OPTIONS`
contains only `temperature`, `maxOutputTokens`, and `timeout`. Supported modes
are `generate`, `chat`, `summarize`, `extract`, and `classify`.

Plain `QUERY` returns a singleton list because Ferret's queryable contract is
list-valued. Use `QUERY ONE` for the scalar result that matches the function
API. `QUERY COUNT` and `QUERY EXISTS` are deliberately unsupported because
either modifier would otherwise cause a potentially billable, state-mutating
provider request.

Query payloads are strings. Bind member expressions before querying:

```fql
LET body = email.body
RETURN QUERY ONE body IN model USING summarize
```

Direct member payloads such as `QUERY ONE email.body IN model` are not part of
Ferret's current query-expression grammar.

## Sessions

`SESSION` accepts only a stateless model:

```fql
LET session = AI::LLM::SESSION(model, {
  instructions: "Answer as a concise technical editor.",
  context: {
    mode: "local",
    overflow: "error",
    maxTokens: 128000,
    reserveOutputTokens: 4000
  }
})

LET first = AI::LLM::CHAT(session, "My API returns widgets.")
LET second = AI::LLM::CHAT(session, "What does my API return?")
RETURN {
  first: first,
  second: second,
  history: AI::LLM::HISTORY(session)
}
```

Only local context with `overflow: "error"` is supported in version 1.
`maxTokens` and `reserveOutputTokens` are validated configuration metadata; the
module does not proactively count tokens. Provider context-window failures are
reported as `AI_LLM_CONTEXT_LIMIT_ERROR`.

Sessions replay visible text history on each request. They do not preserve
hidden reasoning, provider-private response items, tools, or server-side
conversation state. A successful call commits all input messages and the
assistant response atomically. Failed, refused, malformed, or schema-invalid
responses do not change history. Structured assistant responses appear in
history as their validated JSON text.

`RESET` clears history, `FORK` creates an independent copy, and `HISTORY`
returns a copy. Sessions created by `MODEL(..., {session: true})`, `SESSION`, or
`FORK` are scoped to the current Ferret execution and are closed after that run,
including sessions nested in arrays or objects.

## Limitations

This initial module is text-only. Streaming, tools, agents, MCP, embeddings,
OCR, vision, vector search, persistent sessions, provider-side conversations,
custom base URLs, retries, and model discovery are outside its version 1
contract.
