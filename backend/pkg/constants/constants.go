package constants

import "time"

const (
	CookieAccessToken  = "goshort_at"
	CookieRefreshToken = "goshort_rt"

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

const (
	RoleUser       = "user"
	RoleAdvertiser = "advertiser"
	RoleAdmin      = "admin"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

const (
	RedisPrefixBlacklist = "bl:"
	RedisPrefixRateLimit = "rl:"
)

const (
	AdStatusActive  = "active"
	AdStatusPaused  = "paused"
	AdStatusDeleted = "deleted"
)

const (
	ExpiresUnitMinutes = "minutes"
	ExpiresUnitHours   = "hours"
	ExpiresUnitDays    = "days"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

const (
	SameSiteLax    = "Lax"
	SameSiteStrict = "Strict"
)

const (
	LocaleCity = "en"
)

const (
	DeviceDesktop = "desktop"
	DeviceMobile  = "mobile"
	DeviceBot     = "bot"
)

const (
	RedisPrefixFaucet    = "faucet:"
	RedisPrefixWeb3Block  = "web3:last_block"
	RedisPrefixQualityIP = "q:ip:"
	RedisPrefixQualityFP = "q:fp:"
)

const (
	DefaultBalance   = "0.00000000"
	FaucetAmount     = 20
	FaucetCooldown   = 24 * time.Hour
	FaucetClaimLimit = 1
	PollInterval     = 15 * time.Second
)

const (
	EmailVerificationTokenTTL = 10 * time.Minute
	EmailVerificationCooldown = 10 * time.Minute
	PasswordResetTokenTTL     = 10 * time.Minute
	PasswordResetCooldown     = 10 * time.Minute
)

const (
	DefaultPage        = 1
	DefaultPerPage     = 5
	ClaimsPerPage      = 10
	WalletDefaultLimit = 50
)

const (
	MaxLinkSlugRetries    = 5
	MaxTopCountries       = 10
	MaxPerPage            = 100
	ExpiryCleanerInterval = 1 * time.Hour
	QRCodeDefaultSize     = 256
	QRCodeMinSize         = 64
	QRCodeMaxSize         = 1024
	DefaultSlugLength     = 6
	DefaultURLCacheTTL    = 24 * time.Hour
)

const (
	AnalyticsWorkerCount = 5
)

const (
	BlockConfirmations  = 12
	DevETHGasLimit      = 21000
	FaucetDeadline      = 15 * time.Minute
	FaucetHeartbeat     = 1 * time.Hour
)

const (
	RatioTolerance   = 0.05
)

const (
	MaxImageSize     = 5 * 1024 * 1024
	MaxVideoSize     = 20 * 1024 * 1024
	MaxImageSizeText = "5MB"
	MaxVideoSizeText = "20MB"
)

const (
	MediaTypeImage = "image/"
	MediaTypeVideo = "video/"
)

const (
	ContentTypePNG  = "image/png"
	ContentTypeJPEG = "image/jpeg"
	ContentTypeWEBP = "image/webp"
	ContentTypeGIF  = "image/gif"
	ContentTypeMP4  = "video/mp4"
	ContentTypeWEBM = "video/webm"
	ContentTypeOGG  = "video/ogg"
)

const (
	TxTypeEarning       = "EARNING"
	TxTypeAdSpend       = "AD_SPEND"
	TxTypeDeposit       = "DEPOSIT"
	TxTypeWithdrawal    = "WITHDRAWAL"
	TxTypeWithdrawalFee = "WITHDRAWAL_FEE"
	TxTypeFaucet        = "FAUCET"
)

const (
	TxStatusPending   = "PENDING"
	TxStatusConfirmed = "CONFIRMED"
	TxStatusFailed    = "FAILED"
)

const (
	AdEventImpression = "IMPRESSION"
	AdEventClick      = "CLICK"
	AdEventCompletion = "COMPLETION"
	AdEventSkip       = "SKIP"
)

const (
	AdTypePopup        = "POPUP"
	AdTypeBanner       = "BANNER"
	AdTypeNative       = "NATIVE"
	AdTypeVideo        = "VIDEO"
	AdTypeInterstitial = "INTERSTITIAL"
)

const (
	QualityScoreDefault = "1.00"
	QualityScoreSkip    = "0.50"
)

const (
	RejectReasonHoneypotHit          = "HONEYPOT_HIT"
	RejectReasonTooFast              = "TOO_FAST"
	RejectReasonNoMouseMovement      = "NO_MOUSE_MOVEMENT"
	RejectReasonDuplicateIP          = "DUPLICATE_IP"
	RejectReasonDuplicateFingerprint = "DUPLICATE_FINGERPRINT"
)

const (
	WSEventWalletUpdate = "WALLET_UPDATE"
)

const (
	PlatformReward = "0.00100000"
)
