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

var (
	// 4xx
	ErrBadRequest                  = NewHTTPError(StatusBadRequest)
	ErrUnauthorized                = NewHTTPError(StatusUnauthorized)
	ErrPaymentRequired             = NewHTTPError(StatusPaymentRequired)
	ErrForbidden                   = NewHTTPError(StatusForbidden)
	ErrNotFound                    = NewHTTPError(StatusNotFound)
	ErrMethodNotAllowed            = NewHTTPError(StatusMethodNotAllowed)
	ErrNotAcceptable               = NewHTTPError(StatusNotAcceptable)
	ErrProxyAuthRequired           = NewHTTPError(StatusProxyAuthRequired)
	ErrRequestTimeout              = NewHTTPError(StatusRequestTimeout)
	ErrConflict                    = NewHTTPError(StatusConflict)
	ErrGone                        = NewHTTPError(StatusGone)
	ErrLengthRequired              = NewHTTPError(StatusLengthRequired)
	ErrPreconditionFailed          = NewHTTPError(StatusPreconditionFailed)
	ErrRequestEntityTooLarge       = NewHTTPError(StatusRequestEntityTooLarge)
	ErrRequestURITooLong           = NewHTTPError(StatusRequestURITooLong)
	ErrUnsupportedMediaType        = NewHTTPError(StatusUnsupportedMediaType)
	ErrRangeNotSatisfiable         = NewHTTPError(StatusRequestedRangeNotSatisfiable)
	ErrExpectationFailed           = NewHTTPError(StatusExpectationFailed)
	ErrTeapot                      = NewHTTPError(StatusTeapot)
	ErrTooManyRequests             = NewHTTPError(StatusTooManyRequests)
	ErrRequestHeaderFieldsTooLarge = NewHTTPError(StatusRequestHeaderFieldsTooLarge)
	ErrUnavailableForLegalReasons  = NewHTTPError(StatusUnavailableForLegalReasons)

	// 5xx
	ErrInternalServer                = NewHTTPError(StatusInternalServerError)
	ErrNotImplemented                = NewHTTPError(StatusNotImplemented)
	ErrBadGateway                    = NewHTTPError(StatusBadGateway)
	ErrServiceUnavailable            = NewHTTPError(StatusServiceUnavailable)
	ErrGatewayTimeout                = NewHTTPError(StatusGatewayTimeout)
	ErrHTTPVersionNotSupported       = NewHTTPError(StatusHTTPVersionNotSupported)
	ErrVariantAlsoNegotiates         = NewHTTPError(StatusVariantAlsoNegotiates)
	ErrInsufficientStorage           = NewHTTPError(StatusInsufficientStorage)
	ErrLoopDetected                  = NewHTTPError(StatusLoopDetected)
	ErrNotExtended                   = NewHTTPError(StatusNotExtended)
	ErrNetworkAuthenticationRequired = NewHTTPError(StatusNetworkAuthenticationRequired)
)

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
