package model

import (
	"bbs/internal/model/DO"
	"bbs/pkg/global"
	"gorm.io/gorm"
	"strings"
)

// SysUser 用户基本信息，用于登录
type SysUser struct {
	OpenId   string //微信OpenID
	UnionId  string //微信UnionID
	Username string //用户名
	Password string //密码
	Email    string //邮箱
	Phone    string //电话号码
	Gender   int8   //性别
	School   string //学校列表
	Avatar   string //头像Url
	Status   int8   //账号状态
	BaseModel
}

const (
	UserFieldName   = "username"
	UserFieldPoint  = "point"
	UserFieldAvatar = "avatar"
	UserFieldSchool = "school"
	UserFieldStatus = "status"
)

// GetRedisUser Redis中要写入的属性
func (user *SysUser) GetRedisUser() []interface{} {
	r := []interface{}{
		UserFieldName, user.Username,
		UserFieldAvatar, user.Avatar,
		UserFieldSchool, user.School,
		UserFieldStatus, user.Status,
	}
	return r
}

func (*SysUser) TableName() string {
	return "sys_user"
}

// AddUser 新建用户
func (user *SysUser) AddUser(tx *gorm.DB) error {
	if err := tx.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserById 根据ID获取用户
func (user *SysUser) GetUserById() error {
	err := global.Db.First(user, user.Id).Error
	return err
}

// UpdateUser 更新用户
func (user *SysUser) UpdateUser() error {
	return global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func (user *SysUser) ToCachedUser() *DO.CachedUser {
	schoolList := strings.Split(user.School, ",")
	u := &DO.CachedUser{
		Username: user.Username,
		Status:   user.Status,
		School:   schoolList,
	}
	return u
}

func (user *SysUser) GetUserInfo() error {
	return nil
}

// DeleteUserByIds 根据传入ID删除用户
func DeleteUserByIds(ids []int64) error {
	return global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&SysUser{}, ids).Error; err != nil {
			return err
		}
		return nil
	})
}
