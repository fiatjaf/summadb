package responses

import log "github.com/Sirupsen/logrus"

func NotFound() Error {
	log.Debug("not_found")
	return Error{"not_found", "missing", 404}
}

func ConflictError() Error {
	log.Debug("conflict error")
	return Error{"conflict", "document update conflict", 409}
}

func Unauthorized() Error {
	log.Debug("unauthorized error")
	return Error{"unauthorized", "you are not authorized to access this db", 401}
}

func BadRequest(msgs ...string) Error {
	msg := "bad request"
	if len(msgs) > 0 {
		msg = msgs[0]
	}
	log.WithField("msg", msg).Debug("bad_request")
	return Error{"bad_request", msg, 400}
}

func UnknownError(cause ...string) Error {
	log.WithField("cause", cause).Debug("unknown_error")
	return Error{"unknown_error", "unknown", 500}
}
