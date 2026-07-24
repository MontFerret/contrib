package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	// ErrorCode identifies a stable AI::LLM failure category.
	ErrorCode string

	// Role is the provider-neutral role of a text message.
	Role string

	// ContentType identifies provider-neutral message content.
	ContentType string

	// Mode identifies an AI::LLM operation.
	Mode string

	// Content is provider-neutral message content. Version 1 supports text only.
	Content struct {
		Type ContentType
		Text string
	}

	// Message is a provider-neutral input or output message.
	Message struct {
		Role    Role
		Content Content
	}

	// Usage contains normalized token counts reported by a provider.
	Usage struct {
		InputTokens  int64
		OutputTokens int64
		TotalTokens  int64
	}

	// ExecutionOptions controls a provider request.
	ExecutionOptions struct {
		Temperature     *float64
		MaxOutputTokens int64
		Timeout         time.Duration
	}

	// Request is a provider-neutral text-generation request.
	Request struct {
		Messages     []Message
		Instructions string
		Options      ExecutionOptions
	}

	// StructuredRequest is a provider-neutral structured-generation request.
	StructuredRequest struct {
		Name        string
		Description string
		Request
		Schema Schema
	}

	// Response is a normalized provider response. Raw contains copied provider JSON.
	Response struct {
		ID      string
		Model   string
		Text    string
		RawJSON json.RawMessage
		Usage   Usage
	}

	// ModelOptions configures a provider-backed stateless model.
	ModelOptions struct {
		Model   string
		APIKey  string
		Session bool
	}

	// ContextOptions configures local-session context metadata.
	ContextOptions struct {
		Mode                string
		Overflow            string
		MaxTokens           int64
		ReserveOutputTokens int64
	}

	// SessionOptions configures a local session.
	SessionOptions struct {
		Instructions string
		Context      ContextOptions
	}

	// SemanticOptions configures an operation without controlling execution policy.
	SemanticOptions struct {
		Schema       Schema
		Instructions string
		Style        string
		Messages     []Message
		Labels       []string
		MaxWords     int64
	}

	// OperationRequest describes one provider-neutral AI operation.
	OperationRequest struct {
		Mode      Mode
		Input     string
		Execution ExecutionOptions
		Semantic  SemanticOptions
	}

	// ModelConfig is retained as the provider-facing name for model options.
	ModelConfig = ModelOptions
	// GenerationRequest is the provider-facing name for a text request.
	GenerationRequest = Request
	// StructuredGenerationRequest is the provider-facing name for a structured request.
	StructuredGenerationRequest = StructuredRequest
	// GenerationResponse is the provider-facing name for a normalized response.
	GenerationResponse = Response

	// Generator executes provider-neutral text generation.
	Generator interface {
		Generate(context.Context, Request) (Response, error)
	}

	// StructuredGenerator executes provider-neutral structured generation.
	StructuredGenerator interface {
		GenerateStructured(context.Context, StructuredRequest) (Response, error)
	}

	// Target can execute every v1 AI::LLM operation.
	Target interface {
		Generator
		StructuredGenerator
	}

	// Model is an immutable, stateless, Ferret-visible provider model.
	Model interface {
		runtime.Value
		runtime.Queryable
		Target
		Provider() string
		ModelName() string
	}

	// Session is a serialized, local, Ferret-visible conversation resource.
	Session interface {
		runtime.Value
		runtime.Queryable
		runtime.Resource
		Target
		Reset() error
		Fork(context.Context) (Session, error)
		History() []Message
	}

	// ProviderFactory creates explicit-credential models for one provider.
	ProviderFactory interface {
		Name() string
		NewModel(context.Context, ModelOptions) (Model, error)
	}
)

const (
	ErrProvider                ErrorCode = "AI_LLM_PROVIDER_ERROR"
	ErrAuth                    ErrorCode = "AI_LLM_AUTH_ERROR"
	ErrRateLimit               ErrorCode = "AI_LLM_RATE_LIMIT_ERROR"
	ErrTimeout                 ErrorCode = "AI_LLM_TIMEOUT_ERROR"
	ErrContextLimit            ErrorCode = "AI_LLM_CONTEXT_LIMIT_ERROR"
	ErrUnsupportedProvider     ErrorCode = "AI_LLM_UNSUPPORTED_PROVIDER"
	ErrUnsupportedOperation    ErrorCode = "AI_LLM_UNSUPPORTED_OPERATION"
	ErrInvalidOptions          ErrorCode = "AI_LLM_INVALID_OPTIONS"
	ErrInvalidSchema           ErrorCode = "AI_LLM_INVALID_SCHEMA"
	ErrInvalidStructuredOutput ErrorCode = "AI_LLM_INVALID_STRUCTURED_OUTPUT"
	ErrSchemaValidation        ErrorCode = "AI_LLM_SCHEMA_VALIDATION_ERROR"
	ErrRefusal                 ErrorCode = "AI_LLM_REFUSAL"
)

const (
	RoleSystem    Role = "system"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

const ContentText ContentType = "text"

const (
	ModeGenerate  Mode = "generate"
	ModeChat      Mode = "chat"
	ModeSummarize Mode = "summarize"
	ModeExtract   Mode = "extract"
	ModeClassify  Mode = "classify"
)
