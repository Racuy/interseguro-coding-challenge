// parses the Authorization header, shared by the handler and the JWT middleware
package bearer

import "strings"

const prefix = "Bearer "

// splits "Bearer <token>" into the token
func Parse(header string) (string, bool) {
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	return strings.TrimPrefix(header, prefix), true
}
