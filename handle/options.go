package handle

import "net/http"

func flag(r *http.Request, name string, expected ...string) bool {
	qs := r.URL.Query()

	shouldbe := "true"
	if len(expected) > 0 {
		shouldbe = expected[0]
	}

	return qs.Get(name) == shouldbe
}

func param(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}
