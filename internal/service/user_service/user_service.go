package user_service

import (
	"bbs/internal/model"
	"bbs/internal/params"
	"bbs/internal/service/point_service"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"bbs/pkg/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type User struct {
	Ip           string
	RegParamName *params.RegParamName
}

func (u *User) Insert() error {
	return nil
}

func (u *User) RegByName() (sysUser *model.SysUser, err error) {
	funcName := "user_service.RegByName"
	var user model.SysUser
	tx := global.Db.Begin()
	defer func() {
		if err != nil {
			global.LOG.Error(err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Where("username = ?", u.RegParamName.Username).First(&user).Error
	if err == nil {
		return nil, errors.Wrapf(constant.ErrorUserExist, funcName)
	}

	uu := model.SysUser{
		Username: u.RegParamName.Username,
		Password: util.HashAndSalt([]byte(u.RegParamName.Password)),
		Status:   model.UserStatusOk,
	}
	//新增用户
	if err = uu.AddUser(tx); err != nil {
		return nil, errors.Wrap(err, funcName)
	}

	//初始化新用户配置
	if err = UserConfigInit(tx, uu.Id); err != nil {
		return nil, errors.Wrap(err, funcName)
	}

	return &uu, err
}

// UserConfigInit 用户配置初始化
func UserConfigInit(tx *gorm.DB, userId int64) (err error) {
	funcName := "user_service.UserConfigInit"
	//1. 初始化积分
	if err = point_service.PointInit(tx, userId); err != nil {
		return errors.Wrap(err, funcName)
	}

	return err
}
