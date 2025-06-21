package zis

import "regexp"

const (
	regEmail = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

func IsEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}
	return regexp.MustCompile(regEmail).MatchString(email)
}
