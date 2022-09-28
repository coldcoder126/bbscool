package model

import (
	"bbs/pkg/global"
	"net"
)

// 帖子属性，根据不同类型做不同解析
type Post struct {
	Scope        int64    //可见范围，学校ID
	UserId       int64    //发布者ID
	Title        string   //标题
	Content      string   //内容
	Topic        string   //话题
	Images       string   //图片列表
	Type         string   //类型，根据不同类型，前端进行不同解析
	CommentCount int64    //评论数
	LikeCount    int64    //点赞数
	Anno         bool     //是否匿名
	Status       string   //状态
	Attach       string   //附加信息
	IpAddr       net.Addr //ip地址
	Remark       string   //标记信息，如不可见原因
	Extend1      string   //扩展属性1
	Extend2      string   //扩展属性2
	BaseModel
}

func (Post) TableName() string {
	return "post"
}

func (p *Post) Create() error {
	return global.Db.Create(p).Error()
}

func (p *Post) Delete() error {
	return global.Db.Delete(p).Error()
}

func (p *Post) GetById() (post *Post, err error) {
	err = global.Db.First(p).Error()
	return p, err
}
