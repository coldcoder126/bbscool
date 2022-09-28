package constant

import "errors"

var (
	ErrorUserExist         = errors.New("用户已存在")
	ErrorUserNotExist      = errors.New("用户不存在")
	ErrorInvalidPassword   = errors.New("密码错误")
	ErrorInvalidToken      = errors.New("token错误")
	ErrorInvalidRfToken    = errors.New("refreshToken错误")
	ErrorInvalidID         = errors.New("ID错误")
	ErrorScoreInsufficient = errors.New("积分不足")
)
