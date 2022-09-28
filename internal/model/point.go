package model

import (
	"bbs/pkg/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// PointBalance 用户积分余额表
type PointBalance struct {
	UserId  int64 //用户ID
	Balance int   //余额
	BaseModel
}

func (PointBalance) TableName() string {
	return "point_balance"
}

// PointRecord 积分表变动记录表
// 积分可能变动：注册、签到、消费、赠予、被赠予。
type PointRecord struct {
	FromId int64  //赠予人ID，如果是系统，默认为1
	ToId   int64  //被赠予人ID，如果是系统，默认为1
	Opt    string //操作名
	Change int    //积分变动值
	BaseModel
}

func (PointRecord) TableName() string {
	return "point_record"
}

func (pb *PointBalance) UpdateDecr(tx *gorm.DB, userId int64, score int) (err error) {
	fName := "point.Update"
	pb.UserId = userId
	if err = tx.Where("user_id", userId).First(pb).Error(); err != nil {
		return errors.Wrap(err, fName)
	}
	if pb.Balance < score {
		return constant.ErrorScoreInsufficient
	}
	if err = tx.Model(pb).Update("balance", pb.Balance-score).Error(); err != nil {
		return errors.Wrap(err, fName)
	}
	return nil
}

func (pr *PointRecord) Insert(tx *gorm.DB) (err error) {
	return tx.Create(pr).Error()
}
