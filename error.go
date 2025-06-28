package zeno

// HTTPError is returned by handlers and middleware to indicate
// an HTTP failure.  Implementations carry both the underlying error
// message and the HTTP status code that should be sent to the client.
type HTTPError interface {
	error
	StatusCode() int
}

// httpError is the canonical implementation of HTTPError.
type httpError struct {
	Status  int    `json:"status" xml:"status"` // HTTP status code
	Message string `json:"message" xml:"message"`
}

// NewHTTPError returns an HTTPError with the supplied status code and
// optional message.  When msg is empty, the description returned by
// StatusText is used.
//
//	example := NewHTTPError(StatusBadRequest)                     // “Bad Request”
//	custom  := NewHTTPError(StatusBadRequest, "invalid payload")  // custom text
func NewHTTPError(status int, msg ...string) HTTPError {
	m := StatusMessage(status) // fallback description
	if len(msg) > 0 && msg[0] != "" {
		m = msg[0]
	}
	return &httpError{Status: status, Message: m}
}

// Error implements the built‑in error interface.
func (e *httpError) Error() string { return e.Message }

// StatusCode returns the HTTP status code supplied when the error
// was created.
func (e *httpError) StatusCode() int { return e.Status }

// choose returns a predefined “zero‑alloc” error when msg is omitted
// or a brand‑new HTTPError carrying the custom text when provided.
func choose(def HTTPError, status int, msg ...string) HTTPError {
	if len(msg) == 0 {
		return def
	}
	return NewHTTPError(status, msg[0])
}

//
// ---------------------------------------------------------------------------
// Pre‑allocated default errors (zero allocations per use)
// ---------------------------------------------------------------------------
//

var (
	// 4xx
	DefaultBadRequest                  = NewHTTPError(StatusBadRequest)
	DefaultUnauthorized                = NewHTTPError(StatusUnauthorized)
	DefaultPaymentRequired             = NewHTTPError(StatusPaymentRequired)
	DefaultForbidden                   = NewHTTPError(StatusForbidden)
	DefaultNotFound                    = NewHTTPError(StatusNotFound)
	DefaultMethodNotAllowed            = NewHTTPError(StatusMethodNotAllowed)
	DefaultNotAcceptable               = NewHTTPError(StatusNotAcceptable)
	DefaultProxyAuthRequired           = NewHTTPError(StatusProxyAuthRequired)
	DefaultRequestTimeout              = NewHTTPError(StatusRequestTimeout)
	DefaultConflict                    = NewHTTPError(StatusConflict)
	DefaultGone                        = NewHTTPError(StatusGone)
	DefaultLengthRequired              = NewHTTPError(StatusLengthRequired)
	DefaultPreconditionFailed          = NewHTTPError(StatusPreconditionFailed)
	DefaultRequestEntityTooLarge       = NewHTTPError(StatusRequestEntityTooLarge)
	DefaultRequestURITooLong           = NewHTTPError(StatusRequestURITooLong)
	DefaultUnsupportedMediaType        = NewHTTPError(StatusUnsupportedMediaType)
	DefaultRangeNotSatisfiable         = NewHTTPError(StatusRequestedRangeNotSatisfiable)
	DefaultExpectationFailed           = NewHTTPError(StatusExpectationFailed)
	DefaultTeapot                      = NewHTTPError(StatusTeapot)
	DefaultTooManyRequests             = NewHTTPError(StatusTooManyRequests)
	DefaultRequestHeaderFieldsTooLarge = NewHTTPError(StatusRequestHeaderFieldsTooLarge)
	DefaultUnavailableForLegalReasons  = NewHTTPError(StatusUnavailableForLegalReasons)

	// 5xx
	DefaultInternalServerError           = NewHTTPError(StatusInternalServerError)
	DefaultNotImplemented                = NewHTTPError(StatusNotImplemented)
	DefaultBadGateway                    = NewHTTPError(StatusBadGateway)
	DefaultServiceUnavailable            = NewHTTPError(StatusServiceUnavailable)
	DefaultGatewayTimeout                = NewHTTPError(StatusGatewayTimeout)
	DefaultHTTPVersionNotSupported       = NewHTTPError(StatusHTTPVersionNotSupported)
	DefaultVariantAlsoNegotiates         = NewHTTPError(StatusVariantAlsoNegotiates)
	DefaultInsufficientStorage           = NewHTTPError(StatusInsufficientStorage)
	DefaultLoopDetected                  = NewHTTPError(StatusLoopDetected)
	DefaultNotExtended                   = NewHTTPError(StatusNotExtended)
	DefaultNetworkAuthenticationRequired = NewHTTPError(StatusNetworkAuthenticationRequired)
)

//
// ---------------------------------------------------------------------------
// Convenience constructors (optionally supply custom messages)
// ---------------------------------------------------------------------------
//

// 4xx
func ErrBadRequest(msg ...string) HTTPError {
	return choose(DefaultBadRequest, StatusBadRequest, msg...)
}

func ErrUnauthorized(msg ...string) HTTPError {
	return choose(DefaultUnauthorized, StatusUnauthorized, msg...)
}

func ErrPaymentRequired(msg ...string) HTTPError {
	return choose(DefaultPaymentRequired, StatusPaymentRequired, msg...)
}
func ErrForbidden(msg ...string) HTTPError { return choose(DefaultForbidden, StatusForbidden, msg...) }
func ErrNotFound(msg ...string) HTTPError  { return choose(DefaultNotFound, StatusNotFound, msg...) }
func ErrMethodNotAllowed(msg ...string) HTTPError {
	return choose(DefaultMethodNotAllowed, StatusMethodNotAllowed, msg...)
}

func ErrNotAcceptable(msg ...string) HTTPError {
	return choose(DefaultNotAcceptable, StatusNotAcceptable, msg...)
}

func ErrProxyAuthRequired(msg ...string) HTTPError {
	return choose(DefaultProxyAuthRequired, StatusProxyAuthRequired, msg...)
}

func ErrRequestTimeout(msg ...string) HTTPError {
	return choose(DefaultRequestTimeout, StatusRequestTimeout, msg...)
}
func ErrConflict(msg ...string) HTTPError { return choose(DefaultConflict, StatusConflict, msg...) }
func ErrGone(msg ...string) HTTPError     { return choose(DefaultGone, StatusGone, msg...) }
func ErrLengthRequired(msg ...string) HTTPError {
	return choose(DefaultLengthRequired, StatusLengthRequired, msg...)
}

func ErrPreconditionFailed(msg ...string) HTTPError {
	return choose(DefaultPreconditionFailed, StatusPreconditionFailed, msg...)
}

func ErrRequestEntityTooLarge(msg ...string) HTTPError {
	return choose(DefaultRequestEntityTooLarge, StatusRequestEntityTooLarge, msg...)
}

func ErrRequestURITooLong(msg ...string) HTTPError {
	return choose(DefaultRequestURITooLong, StatusRequestURITooLong, msg...)
}

func ErrUnsupportedMediaType(msg ...string) HTTPError {
	return choose(DefaultUnsupportedMediaType, StatusUnsupportedMediaType, msg...)
}

func ErrRangeNotSatisfiable(msg ...string) HTTPError {
	return choose(DefaultRangeNotSatisfiable, StatusRequestedRangeNotSatisfiable, msg...)
}

func ErrExpectationFailed(msg ...string) HTTPError {
	return choose(DefaultExpectationFailed, StatusExpectationFailed, msg...)
}
func ErrTeapot(msg ...string) HTTPError { return choose(DefaultTeapot, StatusTeapot, msg...) }
func ErrTooManyRequests(msg ...string) HTTPError {
	return choose(DefaultTooManyRequests, StatusTooManyRequests, msg...)
}

func ErrRequestHeaderFieldsTooLarge(msg ...string) HTTPError {
	return choose(DefaultRequestHeaderFieldsTooLarge, StatusRequestHeaderFieldsTooLarge, msg...)
}

func ErrUnavailableForLegalReasons(msg ...string) HTTPError {
	return choose(DefaultUnavailableForLegalReasons, StatusUnavailableForLegalReasons, msg...)
}

// 5xx
func ErrInternalServer(msg ...string) HTTPError {
	return choose(DefaultInternalServerError, StatusInternalServerError, msg...)
}

func ErrNotImplemented(msg ...string) HTTPError {
	return choose(DefaultNotImplemented, StatusNotImplemented, msg...)
}

func ErrBadGateway(msg ...string) HTTPError {
	return choose(DefaultBadGateway, StatusBadGateway, msg...)
}

func ErrServiceUnavailable(msg ...string) HTTPError {
	return choose(DefaultServiceUnavailable, StatusServiceUnavailable, msg...)
}

func ErrGatewayTimeout(msg ...string) HTTPError {
	return choose(DefaultGatewayTimeout, StatusGatewayTimeout, msg...)
}

func ErrHTTPVersionNotSupported(msg ...string) HTTPError {
	return choose(DefaultHTTPVersionNotSupported, StatusHTTPVersionNotSupported, msg...)
}

func ErrVariantAlsoNegotiates(msg ...string) HTTPError {
	return choose(DefaultVariantAlsoNegotiates, StatusVariantAlsoNegotiates, msg...)
}

func ErrInsufficientStorage(msg ...string) HTTPError {
	return choose(DefaultInsufficientStorage, StatusInsufficientStorage, msg...)
}

func ErrLoopDetected(msg ...string) HTTPError {
	return choose(DefaultLoopDetected, StatusLoopDetected, msg...)
}

func ErrNotExtended(msg ...string) HTTPError {
	return choose(DefaultNotExtended, StatusNotExtended, msg...)
}

func ErrNetworkAuthenticationRequired(msg ...string) HTTPError {
	return choose(DefaultNetworkAuthenticationRequired, StatusNetworkAuthenticationRequired, msg...)
}

//
// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
//

// IsHTTPError reports whether err conforms to the HTTPError interface.
func IsHTTPError(err error) bool {
	_, ok := err.(HTTPError)
	return ok
}

// ToHTTPError converts any error to an HTTPError.  If err already
// implements HTTPError it is returned unmodified.  Otherwise a new
// HTTPError with StatusInternalServerError is created whose message
// is err.Error().
func ToHTTPError(err error) HTTPError {
	if err == nil {
		return nil
	}
	if e, ok := err.(HTTPError); ok {
		return e
	}
	return NewHTTPError(StatusInternalServerError, err.Error())
}
