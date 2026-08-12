package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/config"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/middleware"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	"github.com/starnet/crm/pkg/logger"
	"github.com/starnet/crm/pkg/wecom"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler 认证处理器
type AuthHandler struct{}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login 独立模式登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "请输入用户名和密码")
		return
	}

	var user model.User
	if err := repomysql.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		dto.Unauthorized(c, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		dto.Unauthorized(c, "用户名或密码错误")
		return
	}

	token, err := h.generateToken(&user)
	if err != nil {
		dto.InternalError(c, "生成令牌失败")
		return
	}

	dto.Success(c, dto.LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Name:     user.Name,
		Role:     string(user.Role),
		TenantID: user.TenantID,
	})
}

// Register 独立模式注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var exist model.User
	if err := repomysql.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		dto.BadRequest(c, "用户名已被注册")
		return
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		dto.InternalError(c, "密码加密失败")
		return
	}

	// 获取默认租户
	tenant, err := repomysql.SeedDefaultTenant()
	if err != nil {
		dto.InternalError(c, "创建租户失败: "+err.Error())
		return
	}

	user := model.User{
		ID:           uuid.New().String(),
		TenantID:     tenant.ID,
		Username:     req.Username,
		PasswordHash: string(hash),
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		Role:         model.RoleMember,
	}

	if err := repomysql.DB.Create(&user).Error; err != nil {
		dto.InternalError(c, "注册失败: "+err.Error())
		return
	}

	token, err := h.generateToken(&user)
	if err != nil {
		dto.InternalError(c, "生成令牌失败")
		return
	}

	dto.Success(c, dto.LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Name:     user.Name,
		Role:     string(user.Role),
		TenantID: user.TenantID,
	})
}

// WecomAuth 企微OAuth登录（前端回调后调用）
func (h *AuthHandler) WecomAuth(c *gin.Context) {
	var req dto.WecomAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败")
		return
	}

	cfg := config.AppConfig.WeCom
	if cfg.CorpID == "" || cfg.Secret == "" {
		dto.InternalError(c, "企微配置未完成")
		return
	}

	// 1. 获取access_token
	accessTokenResp, err := wecom.GetAccessToken(cfg.CorpID, cfg.Secret)
	if err != nil {
		logger.Log.Error("获取企微access_token失败", zap.Error(err))
		dto.InternalError(c, "企微认证失败")
		return
	}

	// 2. 通过code获取UserID
	userInfoResp, err := wecom.GetUserIDByCode(accessTokenResp.AccessToken, req.Code)
	if err != nil {
		logger.Log.Error("获取企微UserID失败", zap.Error(err))
		dto.InternalError(c, "获取企微用户信息失败")
		return
	}

	// 3. 获取用户详细信息
	userDetail, err := wecom.GetUserDetail(accessTokenResp.AccessToken, userInfoResp.UserID)
	if err != nil {
		logger.Log.Error("获取企微用户详情失败", zap.Error(err))
		dto.InternalError(c, "获取企微用户详情失败")
		return
	}

	// 4. 查找或创建用户
	var user model.User
	err = repomysql.DB.Where("wecom_user_id = ?", userDetail.UserID).First(&user).Error
	if err != nil {
		// 用户不存在，创建新用户
		// 获取默认租户
		tenant, tenantErr := repomysql.SeedDefaultTenant()
		if tenantErr != nil {
			dto.InternalError(c, "创建租户失败")
			return
		}

		deptName := ""
		if len(userDetail.Department) > 0 {
			deptName = wecom.GetDepartmentName(accessTokenResp.AccessToken, userDetail.Department[0])
		}

		wecomUID := userDetail.UserID
		user = model.User{
			ID:          uuid.New().String(),
			TenantID:    tenant.ID,
			WecomUserID: &wecomUID,
			Name:        userDetail.Name,
			Phone:       userDetail.Mobile,
			Email:       userDetail.Email,
			Department:  deptName,
			Avatar:      userDetail.Avatar,
			Role:        model.RoleMember,
		}

		if err := repomysql.DB.Create(&user).Error; err != nil {
			logger.Log.Error("创建企微用户失败", zap.Error(err))
			dto.InternalError(c, "创建用户失败")
			return
		}
	} else {
		// 更新用户信息
		updates := map[string]interface{}{
			"name":  userDetail.Name,
			"phone": userDetail.Mobile,
			"email": userDetail.Email,
			"avatar": userDetail.Avatar,
		}
		if len(userDetail.Department) > 0 {
			updates["department"] = wecom.GetDepartmentName(accessTokenResp.AccessToken, userDetail.Department[0])
		}
		repomysql.DB.Model(&user).Updates(updates)
	}

	// 5. 生成JWT
	token, err := h.generateToken(&user)
	if err != nil {
		dto.InternalError(c, "生成令牌失败")
		return
	}

	dto.Success(c, dto.LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Name:     user.Name,
		Role:     string(user.Role),
		TenantID: user.TenantID,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	dto.Success(c, gin.H{"message": "登出成功"})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	var user model.User
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
		dto.NotFound(c, "用户不存在")
		return
	}

	dto.Success(c, user)
}

// generateToken 生成JWT令牌
func (h *AuthHandler) generateToken(user *model.User) (string, error) {
	claims := &middleware.Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.AppConfig.JWT.ExpireH) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "starnet-crm",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

// WecomVerify 企微回调URL验证（GET）
func (h *AuthHandler) WecomVerify(c *gin.Context) {
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if echostr == "" {
		dto.BadRequest(c, "缺少echostr参数")
		return
	}

	cfg := config.AppConfig.WeCom
	if cfg.Token == "" || cfg.EncodingAESKey == "" {
		// 配置未完成时回显原始值（仅用于开发调试）
		logger.Log.Warn("企微Token/EncodingAESKey未配置，回显原始echostr")
		c.String(http.StatusOK, echostr)
		return
	}

	// 签名校验 + AES解密
	plaintext, err := wecom.VerifyURL(msgSignature, timestamp, nonce, echostr, cfg.Token, cfg.EncodingAESKey)
	if err != nil {
		logger.Log.Error("企微URL验证失败", zap.Error(err))
		c.String(http.StatusForbidden, "verify failed")
		return
	}

	logger.Log.Info("企微URL验证成功")
	c.String(http.StatusOK, plaintext)
}

// WecomOAuthConfig 返回企微OAuth配置（供前端构建授权URL）
func (h *AuthHandler) WecomOAuthConfig(c *gin.Context) {
	cfg := config.AppConfig.WeCom
	dto.Success(c, gin.H{
		"corp_id":      cfg.CorpID,
		"redirect_uri": cfg.RedirectURI,
	})
}
