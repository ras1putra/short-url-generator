package constants

import "time"

const (
	CookieAccessToken  = "goshort_at"
	CookieRefreshToken = "goshort_rt"

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)
