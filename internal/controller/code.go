package controller

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy
	CodeNeedAuth
	CodeInvalidToken
	CodeNeedLogin
)

var codeMessageMap = map[ResCode]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeServerBusy:      "服务繁忙",
	CodeNeedAuth:        "需要登录",
	CodeInvalidToken:    "无效token",
	CodeNeedLogin:       "需要登录",
}

func (c ResCode) Message() string {
	message, ok := codeMessageMap[c]
	if !ok {
		message = codeMessageMap[CodeServerBusy]
	}
	return message
}
