package trigger

import "strings"

func convertValueToTopic(str string, typ string) string {
	parts := strings.Split(str, "/")

	fullTopic := make([]string, len(parts)+1)
	fullTopic[0] = parts[0]
	fullTopic[1] = typ
	copy(fullTopic[2:], parts[1:])

	return strings.Join(fullTopic, "/")
}
