package model

const (
	SchoolJunior  = iota //初中
	SchoolSenior         //高中
	SchoolCollage        //大学
)

type School struct {
	FullName     string  //全名
	Level        int8    //类型
	Label        string  //标签
	AbbrEn       string  //英文简称
	AbbrZh       string  //中文简称
	Badge        string  //校徽
	Province     string  //省份
	City         string  // 城市
	Friend       []int64 //盟校
	MainColor    string  //主色
	BaseColor    string  //基色
	Open         bool    //是否对外开放，默认开放
	OutsideRead  bool    //对外区域可读，即为开放外部
	OutsideWrite bool    //对外区域可写，前提必须可读
	BaseModel
}
