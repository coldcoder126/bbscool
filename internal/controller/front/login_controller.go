package front

import (
	"bbs/internal/controller"
	"bbs/internal/params"
	"bbs/internal/service/login_service"
	"bbs/pkg/app"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"bbs/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type LoginController struct {
}

//todo
//Register 注册
func (e *LoginController) RegisterByWechat(c *gin.Context) {
	return
}

// RegisterByUsername 用用户名注册
func (e *LoginController) RegisterByUsername(c *gin.Context) {
	p := new(params.RegParamName)
	// 1. 绑定和校验参数
	if err := app.BindAndValidate(c, p); err != nil {
		app.Response(c, http.StatusOK, err.Error(), nil)
		return
	}

	// 2.注册成功返回jwt
	token, refreshToken, err := login_service.RegisterByUsername(p)
	if err != nil {
		global.LOG.Error(err)
		if errors.Is(err, constant.ErrorInvalidToken) {
			//如果是token生成错误，提示重新登录
			app.Response(c, http.StatusBadRequest, controller.CodeNeedLogin, nil)
		} else {
			app.Response(c, http.StatusBadRequest, controller.CodeServerBusy, nil)
		}
		return
	}

	app.ResponseOk(c, gin.H{
		"token":        token,
		"refreshToken": refreshToken,
		"expire_time":  time.Now().Add(jwt.TtlToken).Unix(),
	})
}

// LoginByUsername 使用用户名登录
func (e *LoginController) LoginByUsername(c *gin.Context) {
	var p params.RegParamName
	if err := app.BindAndValidate(c, &p); err != nil {
		global.LOG.Error(err)
		app.Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	user, err := login_service.LoginByUsername(&p)
	if err != nil {
		global.LOG.Error(err)
		app.Response(c, http.StatusOK, controller.CodeInvalidPassword, nil)
		return
	}

	token, refreshToken, err := jwt.GenerateToken(user)
	if err != nil {
		global.LOG.Error(err)
		app.Response(c, http.StatusOK, controller.CodeServerBusy, nil)
		return
	}

	app.ResponseOk(c, gin.H{
		"token":        token,
		"refreshToken": refreshToken,
		"expire_time":  time.Now().Add(jwt.TtlToken).Unix(),
	})
}

// LoginByRefreshToken 使用refreshToken登录，并刷新两个token
func (e *LoginController) LoginByRefreshToken(c *gin.Context) {
	var rt params.RefreshToken
	if err := app.BindAndValidate(c, &rt); err != nil {
		global.LOG.Error(err)
		app.Response(c, http.StatusBadRequest, controller.CodeInvalidToken, nil)
		return
	}

	token, refreshToken, err := login_service.LoginByRefreshToken(rt.RefreshToken)

	if err != nil {
		global.LOG.Error(err)
		app.Response(c, http.StatusOK, controller.CodeServerBusy, nil)
		return
	}

	app.ResponseOk(c, gin.H{
		"token":        token,
		"refreshToken": refreshToken,
		"expire_time":  time.Now().Add(jwt.TtlToken).Unix(),
	})
}
