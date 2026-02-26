package service

import (
	"errors"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
)

type RefreshToken struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

var tokenMap = sync.Map{} //make(map[string]*RefreshToken)
var userMap = sync.Map{}  //make(map[string]*RefreshToken)

func NewRefreshToken(username, token string) *RefreshToken {
	refreshToken := &RefreshToken{
		Username:  username,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	clear(username)
	// tokenMap[token] = refreshToken
	// userMap[username] = refreshToken
	tokenMap.Store(token, refreshToken)
	userMap.Store(username, refreshToken)
	return refreshToken
}
func clear(username string) {
	val, ok := userMap.LoadAndDelete(username)
	if ok {
		tokenMap.Delete(val.(*RefreshToken).Token)
	}
	// if refreshToken, exists := refreshToken[username]; exists {
	// 	delete(tokenMap, refreshToken.Token)
	// 	delete(userMap, refreshToken.Username)
	// }
}
func GetRefreshToken(userName string) *RefreshToken {
	return NewRefreshToken(userName, helper.RandomString(32))
}

func FindUsernameByToken(token string) (string, error) {
	val, ok := tokenMap.Load(token)
	if ok {
		token := val.(*RefreshToken)
		if time.Now().After(token.ExpiresAt) {
			clear(token.Username)
			tokenMap.Delete(token)
			return "", errors.New("token expired")
		}

		// tokenMap.Delete(val.(*RefreshToken).Token)
	}
	// if refreshToken, exists := tokenMap[token]; exists {
	// 	if time.Now().After(refreshToken.ExpiresAt) {
	// 		delete(tokenMap, token)
	// 		return "", errors.New("token expired")
	// 	}
	// 	return refreshToken.Username, nil
	// }
	return "", errors.New("token not found")
}
