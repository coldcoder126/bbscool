package constant

const (
	//全局
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400
	FAIL_ADD_DATA  = 800

	//用户模块
	ERROR_EXIST_USER     = 10001
	ERROR_NOT_EXIST_USER = 10002
	ERROR_PASS_USER      = 10003
	ERROR_CAPTCHA_USER   = 10004
	FAIL_LOGOUT_USER     = 10005

	//token模块
	ERROR_AUTH_CHECK_TOKEN_FAIL    = 20001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 20002
	ERROR_AUTH_TOKEN               = 20003
	ERROR_AUTH                     = 20004
	ERROR_AUTH_CHECK_FAIL          = 20005

	//上传模块
	ERROR_UPLOAD_SAVE_IMAGE_FAIL    = 30001
	ERROR_UPLOAD_CHECK_IMAGE_FAIL   = 30002
	ERROR_UPLOAD_CHECK_IMAGE_FORMAT = 30003

	//商品模块
	ERROR_NOT_EXIST_PRODUCT = 40002

	//订单模块
	ERROR_NOT_EXIST_ORDER = 50002
)
