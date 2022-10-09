package post_service

import (
	"bbs/internal/model"
	"bbs/internal/model/DO"
	"bbs/internal/service/point_service"
	"bbs/internal/service/user_service"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"encoding/json"
	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

const (
	SysID = iota + 1

	QueLen = 50
	SetLen = 50
)

// AddOne 新增一条
func AddOne(post *model.Post, userId int64) (err error) {
	fName := "model.AddOne"
	ok, err := CheckAuth(post.Scope, userId)
	if err != nil || !ok {
		return errors.Wrapf(err, fName+"CheckAuth failed")
	}

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
	} else {
		// 公共范围贴

	}

	return nil
}

// GetScopePostOrderByUpdateTime 根据范围获取帖子列表，按最新互动时间排序
func GetScopePostOrderByUpdateTime(scope, userId, pageNum, lastScore int64) (posts *[]model.Post, lScore int64, err error) {
	fName := "post_service.GetScopePostOrderByUpdateTime"
	ok, err := CheckAuth(scope, userId)
	if err != nil || !ok {
		return nil, 0, errors.Wrapf(err, fName+"CheckAuth failed")
	}
	//第一页从redis中取出15条
	if pageNum == 1 {

	}
	//第二页，获取redis中所有比latestScore小的全部数据，如果不够7条（15/2），从MySQL中查找，补齐15条
	if pageNum == 2 {

	}
	// 第三页及以后，按latestScore从MySQL中每次取15条

	return nil, 0, nil
}

// GetScopePostOrderByCreateTime 获取指定区域帖子，按帖子创建时间排序
func GetScopePostOrderByCreateTime(scope, userId, pageNum int64, lastTime time.Time) (posts *[]model.Post, lTime time.Time, err error) {
	fName := "post_service.GetScopePostOrderByCreateTime"
	ok, err := CheckAuth(scope, userId)
	if err != nil || !ok {
		return nil, time.Now(), errors.Wrapf(err, fName+"CheckAuth failed")
	}
	//如果是第一页，从redis的list中获取前15个
	if pageNum == 1 {

	}

	//如果是第二页，获取后15个，并按时间戳过滤，如果不够7条（15/2），从MySQL中查找，补齐15条
	if pageNum == 2 {

	}

	// 如果是第三页及之后，从MySQL中获取

	return nil, time.Now(), nil
}

// GetReCmdPosts 获取首页推荐贴
func GetReCmdPosts(userId int) []model.Post {
	return nil
}

// CheckAuth 详细检查用户是否有发/读帖权限
func CheckAuth(scope int64, userId int64) (ok bool, err error) {
	//1.从缓存中获取用户
	cu, err := user_service.GetCachedUser(userId)
	if err != nil {
		return false, err
	}
	//2.检查用户账号状态
	if cu.Status != model.UserStatusOk {
		return false, constant.ErrorInvalidToken
	}

	//3.检查用户是否有权限
	destScope := strconv.Itoa(int(scope))
	userSchStr, err := user_service.GetCachedUserField(userId, "School")
	if err != nil {
		return false, err
	}
	userSchList := strings.Split(userSchStr, ",")
	userSchSet := mapset.NewSet[string](userSchList)
	if !userSchSet.Contains(destScope) {
		//再检查是否可以在盟校发布
		friends := make([]string, 0)
		for i := 0; i < len(userSchList); i++ {
			id, _ := strconv.Atoi(userSchList[i])
			friendStr, _ := DO.GetSchoolValByFieldName(int64(id), "Friend")
			friends = append(friends, strings.Split(friendStr, ",")...)
		}
		friendSet := mapset.NewSet[string](friends)
		if !friendSet.Contains(destScope) {
			return false, constant.ErrorPublishDeny
		}
	}

	return true, err
}
