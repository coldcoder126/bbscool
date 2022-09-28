package point_service

import (
	"bbs/internal/model"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
)

// 根据不同用户不同行为获取积分
// 积分操作枚举
const (
	PointOptRegister = iota
	PointOptAssign
	PointOptPost
	PointOptComment
	PointOptTopic
)

var PointOptMap = map[int]string{
	PointOptRegister: "register",
	PointOptAssign:   "assign",
	PointOptPost:     "post",
	PointOptComment:  "comment",
	PointOptTopic:    "topic",
}

func Opt(optCode int) string {
	return PointOptMap[optCode]
}

// GetPointPost 发帖积分
func getPointByOpt(uid int64, opt int) int {
	switch opt {
	case PointOptRegister:
		return 500
	case PointOptAssign:
		return 5
	case PointOptPost:
		return 8
	case PointOptComment:
		return 1
	}
	return 0
}

// pointSubInCache 封装redis中的分数减少操作，只需传入用户id和操作即可
func pointSubInCache(redis *cache.Redis, userId int64, opt int) (int64, error) {
	funcName := "cache.DoScoreSub"
	score := getPointByOpt(userId, opt)
	if score <= 0 {
		return 0, errors.New(funcName + " score must >= 0")
	}
	val, err := redis.EvalSha(cache.Lua[cache.PointSub], "POINT_SUB", []string{strconv.Itoa(int(userId)), model.UserFieldPoint}, score)
	if err != nil {
		return 0, errors.WithMessage(err, funcName)
	}
	valInt, ok := val.([]interface{})[0].(interface{}).(int64)
	if !ok {
		return 0, fmt.Errorf("%s parse to  int64 failed ", funcName)
	}
	return valInt, err
}

// PointInit 初始化积分
func PointInit(tx *gorm.DB, userId int64) (err error) {
	funcName := "point_service.PointInit"
	point := getPointByOpt(userId, PointOptRegister)
	pr := model.PointRecord{
		FromId: 1,
		ToId:   userId,
		Opt:    Opt(PointOptRegister),
		Change: point,
	}

	pb := model.PointBalance{
		UserId:    userId,
		Balance:   point,
		BaseModel: model.BaseModel{},
	}
	if err = tx.Create(&pr).Error; err != nil {
		return errors.Wrap(err, funcName)
	}

	if err = tx.Create(&pb).Error; err != nil {
		return errors.Wrap(err, funcName)
	}
	return nil
}

// PointSubInDb 数据库中减分数操作
func pointSubInDb(tx *gorm.DB, fromId, toId int64, optCode int) (err error) {
	fName := "point_service."
	score := getPointByOpt(fromId, optCode)
	opt := Opt(optCode)

	// 查出积分余额
	pb := &model.PointBalance{UserId: fromId}
	if err = pb.UpdateDecr(tx, fromId, score); err != nil {
		return errors.Wrap(err, fName)
	}

	pr := &model.PointRecord{FromId: fromId, ToId: toId, Opt: opt, Change: score}
	if err = pr.Insert(tx); err != nil {
		return errors.Wrap(err, fName)
	}
	return nil
}

func PointSub(redis *cache.Redis, tx *gorm.DB, fromId, toId int64, optCode int) (err error) {
	if err = pointSubInDb(tx, fromId, toId, optCode); err != nil {
		return err
	}

	// 更新用户redis中的积分信息
	// todo 如果mysql和redis保持一致性，是不是就不用检查redis中的分数了
	val, err := pointSubInCache(redis, fromId, optCode)
	if err != nil {
		return err
	}
	if val < 0 {
		//积分不足
		return constant.ErrorScoreInsufficient
	}
	return nil
}
