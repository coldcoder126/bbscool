package app

import (
	"bbs/pkg/constant"
	"bbs/pkg/logging"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"net/http"
)

func BindAndValid(c *gin.Context, form interface{}) (int, int) {
	err := c.Bind(form)
	if err != nil {
		logging.Error(err)
		return http.StatusBadRequest, constant.INVALID_PARAMS
	}

	_, err = govalidator.ValidateStruct(form)
	if err != nil {
		logging.Error(err)
		return http.StatusBadRequest, constant.INVALID_PARAMS
	}

	return http.StatusOK, constant.SUCCESS
}

// BindAndValidate 出错直接返回
func BindAndValidate(c *gin.Context, form interface{}) error {
	err := c.Bind(form)
	if err != nil {
		logging.Error(err)
		return err
	}

	_, err = govalidator.ValidateStruct(form)
	if err != nil {
		logging.Error(err)
		return err
	}

	return nil
}
