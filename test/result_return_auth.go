// compile


// Regression test for return err in string! functions.

package auth

import "errors"

var jwtSecret = []byte("secret")

type RefreshClaims struct{}

func VerifyRefreshToken(tokenString string) string! {
	_ = tokenString
	_ = jwtSecret
	return errors.New("invalid token")
}
