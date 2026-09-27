package handlers

import (
	"errors"
	"net/http"
	"time"

	"taste-server/config"
	"taste-server/database"
	"taste-server/middleware"
	"taste-server/models"
	"taste-server/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type registerRequest struct {
	Username string `json:"username" binding:"required,min=2,max=32"`
	Password string `json:"password" binding:"required,min=4,max=32"`
	Nickname string `json:"nickname" binding:"max=32"`
}

// Register 注册并赠送初始尝鲜币（事务保障用户与流水一致）
func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码格式不正确（用户名2-32位，密码4-32位）"})
		return
	}

	var exists int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已被占用"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     nickname,
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return services.BizError("用户名已被占用")
			}
			return err
		}
		u, err := services.WriteCoinTx(tx, user.ID, config.RegisterBonus, 0,
			models.CoinTypeRegister, "user", user.ID, "注册赠送尝鲜币")
		if err != nil {
			return err
		}
		user.CoinBalance = u.CoinBalance
		return nil
	})
	if err != nil {
		code := http.StatusInternalServerError
		msg := "注册失败"
		if biz, ok := services.IsBiz(err); ok {
			code = http.StatusConflict
			msg = biz.Error()
		}
		c.JSON(code, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": issueToken(user.ID),
		"user":  user,
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": issueToken(user.ID),
		"user":  user,
	})
}

// Me 获取当前登录用户
func Me(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func issueToken(uid uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	s, _ := token.SignedString([]byte(config.JWTSecret))
	return s
}
