package middleware

import (
	"bbs/pkg/app"
	"bbs/pkg/constant"
	"bbs/pkg/jwt"
	"bbs/pkg/logging"
	"bbs/pkg/runtime"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
)

const bearerLength = len("Bearer ")

func AppJwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data interface{}

		mytoken := c.Request.Header.Get("Authorization")
		if len(mytoken) < bearerLength {
			app.Response(c, http.StatusUnauthorized, constant.ERROR_AUTH, data)
			c.Abort()
			return
		}
		token := strings.TrimSpace(mytoken[bearerLength:])
		user, err := jwt.ValidateToken(token)
		if err != nil {
			logging.Info(err)
			app.Response(c, http.StatusUnauthorized, constant.ERROR_AUTH_CHECK_TOKEN_FAIL, data)
			c.Abort()
			return
		}
		c.Set(constant.ContextKeyUserObj, user)
		c.Next()
	}
}

// Jwt 1.检查jwt能否解析
func Jwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data interface{}
		url := c.Request.URL.Path

		method := strings.ToLower(c.Request.Method)
		myToken := c.Request.Header.Get("Authorization")
		if len(myToken) < bearerLength {
			app.Response(c, http.StatusUnauthorized, constant.ERROR_AUTH, data)
			c.Abort()
			return
		}
		token := strings.TrimSpace(myToken[bearerLength:])
		user, err := jwt.ValidateToken(token)
		if err != nil {
			logging.Info(err)
			app.Response(c, http.StatusUnauthorized, constant.ERROR_AUTH_CHECK_TOKEN_FAIL, data)
			c.Abort()
			return
		}
		c.Set(constant.ContextKeyUserObj, user)

		//url排除
		//if urlExclude(url) {
		//	c.Next()
		//	return
		//}

		// casbin check
		cb := runtime.Runtime.GetCasbinKey(constant.CASBIN)
		for _, roleName := range user.Roles {
			//超级管理员过滤掉
			if roleName == "admin" {
				break
			}
			logging.Info(roleName, url, method)
			res, err := cb.Enforce(roleName, url, method)
			if err != nil {
				logging.Error(err)
			}
			//logging.Info(res)

			if !res {
				app.Response(c, http.StatusForbidden, constant.ERROR_AUTH_CHECK_FAIL, data)
				c.Abort()
				return
			}
		}
	}
}

//url排除
func urlExclude(url string) bool {
	//公共路由直接放行
	reg := regexp.MustCompile(`[0-9]+`)
	newUrl := reg.ReplaceAllString(url, "*")
	var ignoreUrls = "/admin/menu/build,/admin/user/center,/admin/user/updatePass,/admin/auth/info," +
		"/admin/auth/logout,/admin/materialgroup/*,/admin/material/*,/shop/product/isFormatAttr/*"
	if strings.Contains(ignoreUrls, newUrl) {
		return true
	}

	return false
}
