package handle

func notFound() Error {
	return Error{"not_found", "missing", 404}
}

func conflictError() Error {
	return Error{"conflict", "document update conflict", 409}
}

func badRequest(msgs ...string) Error {
	msg := "bad request"
	if len(msgs) > 0 {
		msg = msgs[0]
	}
	return Error{"bad_request", msg, 400}
}

func unknownError() Error {
	return Error{"unknown_error", "unknown", 500}
}
