package zeno

import "net/http"

// HTTPError represents an HTTP error with HTTP status code and error message
type HTTPError interface {
	error
	// StatusCode returns the HTTP status code of the error
	StatusCode() int
}

// Error contains the error information reported by calling Context.Error().
type httpError struct {
	Status  int    `json:"status" xml:"status"`
	Message string `json:"message" xml:"message"`
}

// NewHTTPError creates a new HttpError instance.
// If the error message is not given, http.StatusText() will be called
// to generate the message based on the status code.
func NewHTTPError(status int, message ...string) HTTPError {
	if len(message) > 0 {
		return &httpError{status, message[0]}
	}
	return &httpError{status, StatusMessage(status)}
}

// Error returns the error message.
func (e *httpError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e *httpError) StatusCode() int {
	return e.Status
}

// 4xx Client Errors
func ErrBadRequest(message ...string) HTTPError {
	return NewHTTPError(http.StatusBadRequest, message...)
} // 400
func ErrUnauthorized(message ...string) HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message...)
} // 401
func ErrPaymentRequired(message ...string) HTTPError {
	return NewHTTPError(http.StatusPaymentRequired, message...)
}                                              // 402
func ErrForbidden(message ...string) HTTPError { return NewHTTPError(http.StatusForbidden, message...) } // 403
func ErrNotFound(message ...string) HTTPError  { return NewHTTPError(http.StatusNotFound, message...) }  // 404
func ErrMethodNotAllowed(message ...string) HTTPError {
	return NewHTTPError(http.StatusMethodNotAllowed, message...)
} // 405
func ErrNotAcceptable(message ...string) HTTPError {
	return NewHTTPError(http.StatusNotAcceptable, message...)
} // 406
func ErrProxyAuthRequired(message ...string) HTTPError {
	return NewHTTPError(http.StatusProxyAuthRequired, message...)
} // 407
func ErrRequestTimeout(message ...string) HTTPError {
	return NewHTTPError(http.StatusRequestTimeout, message...)
}                                             // 408
func ErrConflict(message ...string) HTTPError { return NewHTTPError(http.StatusConflict, message...) } // 409
func ErrGone(message ...string) HTTPError     { return NewHTTPError(http.StatusGone, message...) }     // 410
func ErrLengthRequired(message ...string) HTTPError {
	return NewHTTPError(http.StatusLengthRequired, message...)
} // 411
func ErrPreconditionFailed(message ...string) HTTPError {
	return NewHTTPError(http.StatusPreconditionFailed, message...)
} // 412
func ErrRequestEntityTooLarge(message ...string) HTTPError {
	return NewHTTPError(http.StatusRequestEntityTooLarge, message...)
} // 413
func ErrRequestURITooLong(message ...string) HTTPError {
	return NewHTTPError(http.StatusRequestURITooLong, message...)
} // 414
func ErrUnsupportedMediaType(message ...string) HTTPError {
	return NewHTTPError(http.StatusUnsupportedMediaType, message...)
} // 415
func ErrRangeNotSatisfiable(message ...string) HTTPError {
	return NewHTTPError(http.StatusRequestedRangeNotSatisfiable, message...)
} // 416
func ErrExpectationFailed(message ...string) HTTPError {
	return NewHTTPError(http.StatusExpectationFailed, message...)
}                                           // 417
func ErrTeapot(message ...string) HTTPError { return NewHTTPError(http.StatusTeapot, message...) } // 418
func ErrTooManyRequests(message ...string) HTTPError {
	return NewHTTPError(http.StatusTooManyRequests, message...)
} // 429
func ErrRequestHeaderFieldsTooLarge(message ...string) HTTPError {
	return NewHTTPError(http.StatusRequestHeaderFieldsTooLarge, message...) // 431
}

func ErrUnavailableForLegalReasons(message ...string) HTTPError {
	return NewHTTPError(http.StatusUnavailableForLegalReasons, message...) // 451
}

// 5xx Server Errors
func ErrInternalServer(message ...string) HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message...)
} // 500
func ErrNotImplemented(message ...string) HTTPError {
	return NewHTTPError(http.StatusNotImplemented, message...)
} // 501
func ErrBadGateway(message ...string) HTTPError {
	return NewHTTPError(http.StatusBadGateway, message...)
} // 502
func ErrServiceUnavailable(message ...string) HTTPError {
	return NewHTTPError(http.StatusServiceUnavailable, message...)
} // 503
func ErrGatewayTimeout(message ...string) HTTPError {
	return NewHTTPError(http.StatusGatewayTimeout, message...)
} // 504
func ErrHTTPVersionNotSupported(message ...string) HTTPError {
	return NewHTTPError(http.StatusHTTPVersionNotSupported, message...) // 505
}

func ErrVariantAlsoNegotiates(message ...string) HTTPError {
	return NewHTTPError(http.StatusVariantAlsoNegotiates, message...) // 506
}

func ErrInsufficientStorage(message ...string) HTTPError {
	return NewHTTPError(http.StatusInsufficientStorage, message...) // 507
}

func ErrLoopDetected(message ...string) HTTPError {
	return NewHTTPError(http.StatusLoopDetected, message...) // 508
}

func ErrNotExtended(message ...string) HTTPError {
	return NewHTTPError(http.StatusNotExtended, message...) // 510
}

func ErrNetworkAuthenticationRequired(message ...string) HTTPError {
	return NewHTTPError(http.StatusNetworkAuthenticationRequired, message...) // 511
}
