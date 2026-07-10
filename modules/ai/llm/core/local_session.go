package core

import (
	"context"
	"strings"
	"sync"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// LocalSession serializes provider calls and owns visible text-message history.
type LocalSession struct {
	model        Model
	instructions string
	history      []Message
	context      ContextOptions
	id           uint64
	mu           sync.Mutex
	closed       bool
}

// NewLocalSession creates and tracks a local session backed by a stateless model.
func NewLocalSession(ctx context.Context, model Model, options SessionOptions) (Session, error) {
	if model == nil {
		return nil, NewError(ErrInvalidOptions, "SESSION requires a stateless model")
	}

	options, err := normalizeSessionOptions(options)
	if err != nil {
		return nil, err
	}

	session := &LocalSession{
		model:        model,
		instructions: options.Instructions,
		context:      options.Context,
		id:           newResourceID(),
	}
	if err := TrackSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *LocalSession) Generate(ctx context.Context, request Request) (Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return Response{}, NewError(ErrProvider, "session is closed")
	}

	requestCtx, cancel, err := executionContext(ctx, request.Options)
	if err != nil {
		return Response{}, err
	}
	defer cancel()

	inputs := copyMessages(request.Messages)
	request.Messages = append(copyMessages(s.history), request.Messages...)
	request.Instructions = joinInstructions(s.instructions, request.Instructions)
	response, err := s.model.Generate(requestCtx, request)
	if err != nil {
		return Response{}, normalizeContextError(requestCtx, err)
	}

	s.commit(inputs, response.Text)

	return response, nil
}

func (s *LocalSession) GenerateStructured(ctx context.Context, request StructuredRequest) (Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return Response{}, NewError(ErrProvider, "session is closed")
	}
	if request.Schema.compiled == nil {
		return Response{}, NewError(ErrInvalidSchema, "schema is not compiled")
	}

	requestCtx, cancel, err := executionContext(ctx, request.Options)
	if err != nil {
		return Response{}, err
	}
	defer cancel()

	inputs := copyMessages(request.Messages)
	request.Messages = append(copyMessages(s.history), request.Messages...)
	request.Instructions = joinInstructions(s.instructions, request.Instructions)
	response, err := s.model.GenerateStructured(requestCtx, request)
	if err != nil {
		return Response{}, normalizeContextError(requestCtx, err)
	}
	if _, err := request.Schema.ValidateJSON([]byte(response.Text)); err != nil {
		return Response{}, err
	}

	s.commit(inputs, response.Text)

	return response, nil
}

func (s *LocalSession) executeOperation(ctx context.Context, operation OperationRequest) (runtime.Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return runtime.None, NewError(ErrProvider, "session is closed")
	}

	value, response, inputs, err := executeAgainst(ctx, s.model, operation, s.history, s.instructions)
	if err != nil {
		return runtime.None, OperationError(strings.ToUpper(string(operation.Mode)), err)
	}

	s.commit(inputs, response.Text)

	return value, nil
}

func (s *LocalSession) commit(inputs []Message, assistant string) {
	s.history = append(s.history, copyMessages(inputs)...)
	s.history = append(s.history, TextMessage(RoleAssistant, assistant))
}

func (s *LocalSession) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return NewError(ErrProvider, "session is closed")
	}

	s.history = nil

	return nil
}

func (s *LocalSession) Fork(ctx context.Context) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, NewError(ErrProvider, "session is closed")
	}

	fork := &LocalSession{
		model:        s.model,
		instructions: s.instructions,
		context:      s.context,
		history:      copyMessages(s.history),
		id:           newResourceID(),
	}
	if err := TrackSession(ctx, fork); err != nil {
		return nil, err
	}

	return fork, nil
}

func (s *LocalSession) History() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	return copyMessages(s.history)
}

func (s *LocalSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.history = nil
	s.instructions = ""
	s.model = nil

	return nil
}

func (s *LocalSession) Query(ctx context.Context, query runtime.Query) (runtime.List, error) {
	value, err := ExecuteQuery(ctx, s, query)
	if err != nil {
		return nil, err
	}

	return runtime.NewArrayWith(value), nil
}

func (s *LocalSession) QueryOne(ctx context.Context, query runtime.Query) (runtime.Value, error) {
	return ExecuteQuery(ctx, s, query)
}

func (s *LocalSession) QueryCount(context.Context, runtime.Query) (runtime.Int, error) {
	return runtime.ZeroInt, NewError(ErrUnsupportedOperation, "QUERY COUNT is not supported")
}

func (s *LocalSession) QueryExists(context.Context, runtime.Query) (runtime.Boolean, error) {
	return runtime.False, NewError(ErrUnsupportedOperation, "QUERY EXISTS is not supported")
}

func (s *LocalSession) ResourceID() uint64 {
	return s.id
}

func (s *LocalSession) String() string {
	return commonresource.Display("ai.llm.session")
}

func (s *LocalSession) GoString() string {
	return s.String()
}

func (s *LocalSession) Hash() uint64 {
	return commonresource.Hash("ai.llm.session", s.id)
}

func (s *LocalSession) Copy() runtime.Value {
	return s
}

func (s *LocalSession) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalDisplayJSON("ai.llm.session")
}

func (s *LocalSession) DebugInfo() runtime.DebugInfo {
	return runtime.DebugInfo{TypeName: "ai.llm.session", Display: s.String()}
}
