package DO

import (
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"github.com/pkg/errors"
	"strconv"
)

type CachedSchool struct {
	FullName string  //全名
	AbbrEn   string  //英文简称
	AbbrZh   string  //中文简称
	Province string  //省份
	City     string  // 城市
	Friend   []int64 //盟校
}

// GetSchoolValByFieldName 获取缓存中的属性值
func GetSchoolValByFieldName(sid int64, field string) (string, error) {
	key := constant.RedisPrefixSchool + strconv.Itoa(int(sid))
	val, err := cache.GetRedisClient(cache.DefaultRedisClient).HGet(key, field)
	//所有学校数据都永久放在缓存中
	if err != nil {
		return "", err
	}
	return val, err
}

// GetCachedSchool 获取缓存中的对象
func (cs *CachedSchool) GetCachedSchool(sid int64) error {
	key := constant.RedisPrefixSchool + strconv.Itoa(int(sid))
	val := cache.GetRedisClient(cache.DefaultRedisClient).Get(key)
	cs, ok := val.(*CachedSchool)
	if ok {
		return nil
	}
	return errors.New("cached_school cache parse to CachedUser failed")
}
