package post_service

import (
	"bbs/internal/model"
	"bbs/internal/service/point_service"
	"bbs/pkg/cache"
	"bbs/pkg/global"
	"github.com/pkg/errors"
)

const (
	SysID = iota + 1
)

// AddOne 新增一条
func AddOne(post *model.Post, userId int64) (err error) {
	fName := "model.AddOne"

	post.UserId = userId
	tx := global.Db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 更新用户mysql中的积分信息

	err = point_service.PointSub(cache.GetRedisClient(cache.DefaultRedisClient), tx, userId, int64(SysID), point_service.PointOptPost)
	if err != nil {
		return errors.Wrapf(err, "%s point decrease failed ", fName)
	}

	// 创建帖子
	if err = tx.Create(post).Error; err != nil {
		return errors.Wrapf(err, "%s post publish failed", fName)
	}

	// 按发布时间，将帖子放入redis队列中

	// todo 1.将post信息放入队列，消费后写入ES
	return nil
}

// GetPostsByScope 根据范围获取帖子列表，默认按最新互动时间排序
func GetPostsByScope(scope int64) []model.Post {
	//优先从redis中取
	return nil
}

// GetReCmdPosts 获取首页推荐贴
func GetReCmdPosts(userId int) []model.Post {
	return nil
}
