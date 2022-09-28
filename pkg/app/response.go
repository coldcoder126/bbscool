package app

import (
	"bbs/pkg/constant"
	"github.com/gin-gonic/gin"
	"net/http"
)

type R struct {
	Code int         `json:"status"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type RPage struct {
	Code      int         `json:"status"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Total     int         `json:"total"`
	TotalPage int         `json:"totalPage"`
}

func Response(c *gin.Context, httpCode int, errCode interface{}, data interface{}) {
	switch errCode.(type) {
	case int:
		intCode := errCode.(int)
		c.JSON(httpCode, R{
			Code: intCode,
			Msg:  constant.GetMsg(intCode),
			Data: data,
		})
	case string:
		strCode := errCode.(string)
		c.JSON(httpCode, R{
			Code: 9999,
			Msg:  strCode,
			Data: data,
		})
	}
}

func ResponseOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, R{
		Code: constant.SUCCESS,
		Msg:  constant.GetMsg(constant.SUCCESS),
		Data: data,
	})
}

func ResponsePage(c *gin.Context, httpCode int, errCode interface{}, data interface{}, total, totalPage int) {
	intCode := errCode.(int)
	c.JSON(httpCode, RPage{
		Code:      intCode,
		Msg:       constant.GetMsg(intCode),
		Data:      data,
		Total:     total,
		TotalPage: totalPage,
	})
	return
}
