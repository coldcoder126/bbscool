package login_service

import (
	"bbs/internal/model"
	"bbs/internal/params"
	"bbs/internal/service/user_service"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"bbs/pkg/jwt"
	"bbs/pkg/util"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func RegisterByUsername(p *params.RegParamName) (token, refreshToken string, err error) {
	fName := "RegisterByUsername"
	// 1. 注册并初始化新用户配置
	userService := user_service.User{RegParamName: p}
	user, err := userService.RegByName()
	if err != nil {
		return "", "", errors.Wrap(err, fName)
	}

	// 2.注册成功返回jwt
	token, refreshToken, err = jwt.GenerateToken(user)
	if err != nil {
		return "", "", errors.Wrap(err, fName)
	}

	return
}

func LoginByRefreshToken(refreshToken string) (string, string, error) {
	fName := "login_service.LoginByRefreshToken"
	// 校验并获取refreshToken中的用户数据
	u, err := jwt.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.Wrap(constant.ErrorInvalidRfToken, "login_service.LoginByRefreshToken")
	}
	cur_signature := strings.Split(refreshToken, ".")[2]

	//校验是否是缓存中的refreshToken，还是已经废弃的refreshToken
	key := constant.RedisPrefixAuth + strconv.Itoa(int(u.Id))
	val, err := cache.GetRedisClient(cache.DefaultRedisClient).GetStr(key)
	if err != nil || val != cur_signature {
		return "", "", errors.Wrap(constant.ErrorInvalidRfToken, fName)
	}

	sysUser := &model.SysUser{}
	sysUser.Id = u.Id
	if err = sysUser.GetUserById(); err != nil {
		return "", "", errors.Wrap(err, fName)
	}

	token, rfToken, err := jwt.GenerateToken(sysUser)
	if err != nil {
		return "", "", errors.Wrap(err, fName)
	}

	return token, rfToken, err
}

func LoginByUsername(p *params.RegParamName) (*model.SysUser, error) {
	var user model.SysUser

	err := global.Db.Where("username = ?", p.Username).First(&user).Error
	if err != nil {
		return nil, constant.ErrorUserNotExist
	}
	if !util.ComparePwd(user.Password, []byte(p.Password)) {
		return nil, constant.ErrorInvalidPassword
	}
	return &user, err
}
