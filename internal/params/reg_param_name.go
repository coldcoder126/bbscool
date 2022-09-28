package params

type RegParamName struct {
	Username string `form:"username" json:"username" valid:"required~用户名不能为空" binding:"required"`
	Password string `form:"password" json:"password" valid:"required~密码不能为空" binding:"required"`
}

type RefreshToken struct {
	RefreshToken string `form:"refreshToken" json:"refreshToken" valid:"required~refreshToken不能为空" binding:"required"`
}
