package post_service

import (
	"bbs/internal/model"
	"bbs/internal/service/point_service"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"encoding/json"
	"github.com/pkg/errors"
	"strconv"
)

const (
	SysID = iota + 1

	QueLen = 50
	SetLen = 50
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
	if err = point_service.PointSubInDb(tx, userId, int64(SysID), point_service.PointOptPost); err != nil {
		return errors.Wrapf(err, "%s point decrease failed in db ", fName)
	}

	// 创建帖子
	if err = tx.Create(post).Error; err != nil {
		return errors.Wrapf(err, "%s post publish failed", fName)
	}

	//数据库都成功了再更新redis
	//更新redis中的分数
	if _, err = point_service.PointSubInCache(cache.GetRedisClient(cache.DefaultRedisClient), userId, point_service.PointOptPost); err != nil {
		return errors.Wrapf(err, " %s redis failed ", fName)
	}

	// 如果是校级范围，将帖子信息写入redis
	// todo 公共级别帖子采用推荐流的方式
	if post.Scope > 100 {
		queKey := constant.RedisPrefixPostQue + strconv.Itoa(int(post.Scope))
		setKey := constant.RedisPrefixPostSet + strconv.Itoa(int(post.Scope))
		postStr, err := json.Marshal(post)
		if err != nil {
			return errors.Wrap(err, "post json failed in "+fName)
		}

		//按发布时间和最新更新时间，将帖子放入redis队列中
		rds := cache.GetRedisClient(cache.DefaultRedisClient)
		rds.EvalSha(cache.Lua[cache.PostToQue], "POST_TO_QUE", []string{queKey}, QueLen, string(postStr))
		rds.EvalSha(cache.Lua[cache.PointSub], "POST_TO_SET", []string{setKey}, SetLen, string(postStr))

		// todo 1.将post信息放入队列，消费后写入ES
	}

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
