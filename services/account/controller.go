package account

import (
	"net/http"

	"lovebox/models"
	"lovebox/pkg/resp"
	"lovebox/services/system"

	"github.com/gin-gonic/gin"
)

type GinController struct {
	AccountSvc *Service
	SystemSvc  *system.Service
}

func NewGinController(accountSvc *Service, systemSvc *system.Service) *GinController {
	return &GinController{
		AccountSvc: accountSvc,
		SystemSvc:  systemSvc,
	}
}

// GetCaptcha 获取验证码
func (ctrl *GinController) GetCaptcha(c *gin.Context) {
	req := &models.GetCaptchaReq{}
	if err := c.ShouldBind(&req); err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	result, err := ctrl.AccountSvc.GetCaptcha(c.Request.Context(), req.Type)
	if err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusOK, &resp.Response{Result: result})
}

// Login 账号登录
func (ctrl *GinController) Login(c *gin.Context) {
	req := &models.LoginReq{}
	if err := c.ShouldBind(&req); err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	token, _, err := ctrl.AccountSvc.Login(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	result := &models.LoginOrRegisterRes{
		Token: token,
	}
	c.JSON(http.StatusOK, &resp.Response{Result: result})
}

// Register 账号注册
func (ctrl *GinController) Register(c *gin.Context) {
	req := &models.RegisterReq{}
	if err := c.ShouldBind(&req); err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	token, _, err := ctrl.AccountSvc.Register(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	result := &models.LoginOrRegisterRes{
		Token: token,
	}
	c.JSON(http.StatusOK, &resp.Response{Result: result})
}

// Info 查询当前登录账号信息
func (ctrl *GinController) Info(c *gin.Context) {
	result, err := ctrl.AccountSvc.Info(
		c.Request.Context(),
		c.GetUint("id"),
	)
	if err != nil {
		_ = c.Error(err).
			SetType(gin.ErrorTypePublic)
		return
	}

	c.JSON(http.StatusOK, &resp.Response{Result: result})
}
