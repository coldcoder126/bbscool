package router

import (
	"bbs/internal/controller/front"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	//r.Use(middleware.Cross())

	// 用户端api（无需鉴权）
	loginController := front.LoginController{}
	loginRouter := r.Group("/login")
	{
		loginRouter.POST("register-by-username", loginController.RegisterByUsername)
		loginRouter.POST("login-by-username", loginController.LoginByUsername)
		loginRouter.POST("login-by-refresh-token", loginController.LoginByRefreshToken)
	}

	// 发布api，需要鉴权
	postController := front.PostController{}

	return r
}
