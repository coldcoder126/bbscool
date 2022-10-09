package constant

import "time"

const (
	ContextKeyUserObj  = "authedUserObj"
	RedisPrefixInfo    = "info:"
	RedisPrefixAuth    = "auth:"
	RedisPrefixPostQue = "que:"
	RedisPrefixPostSet = "set:"
	RedisPrefixSchool  = "school:"
	CASBIN             = "gin-shop"
	WeChatMenu         = "wechat_menus"
	AppRedisPrefixAuth = "app_auth:"
	AppAuthUser        = "app_auth_user:"
	SmsCode            = "sms_code:"
	SmsLength          = 6
	CityList           = "shop-city:"
	OrderInfo          = "order-info:"

	OneDay    = time.Hour * 24
	TwoDays   = time.Hour * 48
	ThreeDays = time.Hour * 72
	FiveDays  = time.Hour * 120
	OneWeek   = time.Hour * 168
	OneMonth  = time.Hour * 720

	PageSize = 15
)
