package consts

var VERSION = "DEV"

const (
	APIRouteSchema = "/api/%s/%s" // INFO: Schema => /api/{version}/{connection_type}
	ConnTypeAdmin  = "admin"
	ConnTypeWorker = "worker"
	StatusError    = "error"
	StatusOK       = "ok"
)

type ErrorCode int

const (
	ErrorCodeNone ErrorCode = iota
	ErrorCodeNotFound
	ErrorCodeConnectionTypeNotFound
	ErrorCodeConnectionTypeNotImplemented
	ErrorCodeServerInternalError
	ErrorCodeServerIllegalArgument
)

func (e ErrorCode) String() string {
	switch e {
	case ErrorCodeNone:
		return "none"
	case ErrorCodeNotFound:
		return "not_found"
	case ErrorCodeConnectionTypeNotFound:
		return "connection_type_not_found"
	case ErrorCodeConnectionTypeNotImplemented:
		return "connection_type_not_implemented"
	case ErrorCodeServerInternalError:
		return "server_internal_error"
	case ErrorCodeServerIllegalArgument:
		return "server_illegal_argument"
	default:
		return "unknown"
	}
}
