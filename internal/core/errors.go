package core

type Error struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

type ErrorKind string

const (
	ErrUpstream   ErrorKind = "upstream"
	ErrTransport  ErrorKind = "transport"
	ErrValidation ErrorKind = "validation"
	ErrInternal   ErrorKind = "internal"
	ErrNotFound   ErrorKind = "not_found"
)

func UpstreamError(message string, cause error) *Error {
	return &Error{Kind: ErrUpstream, Message: message, Cause: cause}
}

func TransportError(message string, cause error) *Error {
	return &Error{Kind: ErrTransport, Message: message, Cause: cause}
}

func ValidationError(message string, cause error) *Error {
	return &Error{Kind: ErrValidation, Message: message, Cause: cause}
}

func InternalError(message string, cause error) *Error {
	return &Error{Kind: ErrInternal, Message: message, Cause: cause}
}

func NotFound(message string, cause error) *Error {
	return &Error{Kind: ErrNotFound, Message: message, Cause: cause}
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) IsRetryable() bool {
	if e.Kind == ErrTransport || e.Kind == ErrUpstream {
		return true
	} else {
		return false
	}
}
