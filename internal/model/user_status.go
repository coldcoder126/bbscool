package model

import "time"

const (
	UserStatusOk = iota
	UserStatusForbid
	UserStatusLock
)

var statusMap = map[int]string{
	UserStatusOk:     "ok",
	UserStatusForbid: "forbid",
	UserStatusLock:   "lock",
}

// UserStatus 用户账号状态
type UserStatus struct {
	Status    int       //用户状态
	Base      int       //基数
	StartTime time.Time //非正常状态的开始时间
	EndTime   time.Time //非正常状态的结束时间
	BaseModel
}

func (UserStatus) TableName() string {
	return "user_status"
}
