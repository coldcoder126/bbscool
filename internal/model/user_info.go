package model

// UserInfo 用户额外信息
type UserInfo struct {
	desc  string //个人描述
	major string //专业
	grade string //年级
	BaseModel
}

func (UserInfo) TableName() string {
	return "user_info"
}
