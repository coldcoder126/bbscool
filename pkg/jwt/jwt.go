package jwt

import (
	"bbs/internal/model"
	"bbs/internal/model/vo"
	"bbs/pkg/cache"
	"bbs/pkg/constant"
	"bbs/pkg/global"
	"bbs/pkg/logging"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

var jwtSecret []byte

const (
	bearerLength = len("Bearer ")
	TtlToken     = time.Hour * 3
	TtlRefresh   = time.Hour * 720
)

var (
	ErrAbsent  = "token absent"  // 令牌不存在
	ErrInvalid = "token invalid" // 令牌无效
	ErrExpired = "token expired" // 令牌过期
	ErrOther   = "other error"   // 其他错误
)

type userStdClaims struct {
	vo.JwtUser
	//*models.User
	jwt.RegisteredClaims
}

func Init() {
	jwtSecret = []byte(global.CONFIG.App.JwtSecret)
}

func GenerateAppToken(m *model.SysUser, d time.Duration) (string, error) {
	m.Password = ""
	//m.Permissions = []string{}
	//expireTime := time.Now().Add(d)
	stdClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
		Issuer:    "bbsAppGo",
	}

	var jwtUser = vo.JwtUser{
		Id:       m.Id,
		Username: m.Username,
	}

	uClaims := userStdClaims{
		RegisteredClaims: stdClaims,
		JwtUser:          jwtUser,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		logging.Error(err)
	}
	//set cache
	var key = constant.AppRedisPrefixAuth + tokenString
	json, _ := json.Marshal(m)
	err = cache.GetRedisClient(cache.DefaultRedisClient).Set(key, json, d)
	if err != nil {
		global.LOG.Error("GenerateAppToken cache set error", err, "key", key)
	}

	return tokenString, err
}

// GenerateToken 返回token和refresh token
func GenerateToken(m *model.SysUser) (string, string, error) {
	fName := "jwt.GenerateToken"
	m.Password = ""
	stdClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TtlToken)),
		Issuer:    "bbsGo",
	}

	var jwtUser = vo.JwtUser{
		Id:       m.Id,
		Username: m.Username,
		Status:   m.Status,
	}

	uClaims := userStdClaims{
		RegisteredClaims: stdClaims,
		JwtUser:          jwtUser,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", errors.Wrap(err, fName)
	}

	// 生成 refresh token
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TtlRefresh)),
		Issuer:    "bbsGo",
	}
	var rUser = vo.JwtUser{
		Id: m.Id,
	}
	rClaims := userStdClaims{
		RegisteredClaims: refreshClaims,
		JwtUser:          rUser,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, rClaims)
	refreshString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", errors.Wrap(err, fName)
	}
	//refresh token放入缓存中
	//用户信息放入缓存中
	authKey := constant.RedisPrefixAuth + strconv.Itoa(int(jwtUser.Id))
	//refresh token只保存签名部分，节省空间
	signature := strings.Split(refreshString, ".")[2]
	infoKey := constant.RedisPrefixInfo + strconv.Itoa(int(jwtUser.Id))
	if err = cache.GetRedisClient(cache.DefaultRedisClient).Set(authKey, signature, TtlRefresh); err != nil {
		return "", "", errors.Wrap(err, fName)
	}
	if err = cache.GetRedisClient(cache.DefaultRedisClient).HMSet(infoKey, m.GetRedisUser(), TtlToken); err != nil {
		return "", "", errors.Wrap(err, fName)
	}
	return tokenString, refreshString, err
}

//返回id
func GetAppUserId(c *gin.Context) (int64, error) {
	u, exist := c.Get(constant.AppAuthUser)
	if !exist {
		return 0, errors.New("can't get user id")
	}
	user, ok := u.(*vo.JwtUser)

	if ok {
		return user.Id, nil
	}
	return 0, errors.New("can't convert to user struct")
}

//返回user
func GetAppUser(c *gin.Context) (*vo.JwtUser, error) {
	u, exist := c.Get(constant.AppAuthUser)
	if !exist {
		return nil, errors.New("can't get user id")
	}
	user, ok := u.(*vo.JwtUser)
	if ok {
		return user, nil
	}
	return nil, errors.New("can't convert to user struct")
}

//返回 detail user
func GetAppDetailUser(c *gin.Context) (*model.SysUser, error) {
	mytoken := c.Request.Header.Get("Authorization")
	if mytoken == "" {
		return nil, errors.New("user not login")
	}
	token := strings.TrimSpace(mytoken[bearerLength:])
	var key = constant.AppRedisPrefixAuth + token
	val, err := cache.GetRedisClient(cache.DefaultRedisClient).GetStr(key)
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]string)
	userMap[key] = val
	jsonStr := userMap[key]
	user := &model.SysUser{}
	err = json.Unmarshal([]byte(jsonStr), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func RemoveAppUser(c *gin.Context) error {
	mytoken := c.Request.Header.Get("Authorization")
	token := strings.TrimSpace(mytoken[bearerLength:])
	var key = constant.AppRedisPrefixAuth + token
	return cache.GetRedisClient(cache.DefaultRedisClient).Delete(key)
}

func ValidateToken(tokenString string) (*vo.JwtUser, error) {
	if tokenString == "" {
		return nil, constant.ErrorInvalidToken
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if token == nil {
		return nil, constant.ErrorInvalidToken
	}
	claims := userStdClaims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrapf(constant.ErrorInvalidToken, "unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		logging.Error(err)
		return nil, err
	}
	return &claims.JwtUser, err

}

//返回id
func GetAdminUserId(c *gin.Context) (int64, error) {
	u, exist := c.Get(constant.ContextKeyUserObj)
	if !exist {
		return 0, errors.New("can't get user id")
	}
	user, ok := u.(*vo.JwtUser)

	if ok {
		return user.Id, nil
	}
	return 0, errors.New("can't convert to user struct")
}

//返回user
func GetAdminUser(c *gin.Context) (*vo.JwtUser, error) {
	u, exist := c.Get(constant.ContextKeyUserObj)
	if !exist {
		return nil, errors.New("can't get user id")
	}
	user, ok := u.(*vo.JwtUser)
	if ok {
		return user, nil
	}
	return nil, errors.New("can't convert to user struct")
}

//返回 detail user
func GetAdminDetailUser(c *gin.Context) *model.SysUser {
	mytoken := c.Request.Header.Get("Authorization")
	token := strings.TrimSpace(mytoken[bearerLength:])
	var key = constant.RedisPrefixAuth + token
	val, err := cache.GetRedisClient(cache.DefaultRedisClient).GetStr(key)
	if err != nil {
		global.LOG.Error("cache error ", err, "key", key, "cmd : Get", "client", cache.DefaultRedisClient)
		return nil
	}
	userMap := make(map[string]string)
	userMap[key] = val
	jsonStr := userMap[key]
	user := &model.SysUser{}
	json.Unmarshal([]byte(jsonStr), user)
	return user
}

func RemoveUser(c *gin.Context) error {
	mytoken := c.Request.Header.Get("Authorization")
	token := strings.TrimSpace(mytoken[bearerLength:])
	var key = constant.RedisPrefixAuth + token
	return cache.GetRedisClient(cache.DefaultRedisClient).Delete(key)
}
