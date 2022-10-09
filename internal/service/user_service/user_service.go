package user_service

import (
	"bbs/internal/model"
	"bbs/internal/model/DO"
	"bbs/internal/params"
	"bbs/internal/service/point_service"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"bbs/pkg/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
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

// GetCachedUserField 根据用户信息获取缓存中的用户属性
func GetCachedUserField(userId int64, field string) (string, error) {
	key := constant.RedisPrefixInfo + strconv.Itoa(int(userId))

	val, err := cache.GetRedisClient(cache.DefaultRedisClient).HGet(key, field)
	//如果该field不存在
	if err != nil {
		if su, err := SetUserToCache(userId); err == nil {
			cu := su.ToCachedUser()
			val = cu.GetValByFieldName(field)
			return val, err
		}
		return val, err
	}
	_, err = cache.GetRedisClient(cache.DefaultRedisClient).Expire(key, constant.ThreeDays)
	return val, err
}

// GetCachedUser 根据用户ID获取缓存中的用户对象
func GetCachedUser(userId int64) (cu *DO.CachedUser, err error) {
	key := constant.RedisPrefixInfo + strconv.Itoa(int(userId))
	val := cache.GetRedisClient(cache.DefaultRedisClient).Get(key)
	//将val转换为CachedUser
	cu, ok := val.(*DO.CachedUser)
	if !ok {
		return nil, err
	}
	return cu, err
}

// SetUserToCache 查出数据库中的数据然后写入redis
func SetUserToCache(userId int64) (user *model.SysUser, err error) {
	key := constant.RedisPrefixInfo + strconv.Itoa(int(userId))
	user.Id = userId
	if err = user.GetUserById(); err != nil {
		return user, err
	}
	if err = cache.GetRedisClient(cache.DefaultRedisClient).HMSet(key, user.GetRedisUser(), constant.ThreeDays); err != nil {
		return user, err
	}
	return user, err
}
