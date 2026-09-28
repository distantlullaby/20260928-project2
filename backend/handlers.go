package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ---------- 视图对象 ----------

type userVO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	AvatarEmoji string `json:"avatar_emoji"`
}

type responseVO struct {
	ID        uint      `json:"id"`
	WishID    uint      `json:"wish_id"`
	User      userVO    `json:"user"`
	Photos    []string  `json:"photos"`
	Comment   string    `json:"comment"`
	Rating    int       `json:"rating"`
	IsAdopted bool      `json:"is_adopted"`
	CreatedAt time.Time `json:"created_at"`
}

type wishVO struct {
	ID             uint         `json:"id"`
	User           userVO       `json:"user"`
	RestaurantName string       `json:"restaurant_name"`
	Address        string       `json:"address"`
	DishList       string       `json:"dish_list"`
	Reason         string       `json:"reason"`
	RewardCoins    int          `json:"reward_coins"`
	Status         string       `json:"status"`
	ResponseCount  int64        `json:"response_count"`
	Responses      []responseVO `json:"responses,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

func toUserVO(u User) userVO {
	return userVO{ID: u.ID, Username: u.Username, Nickname: u.Nickname, AvatarEmoji: u.AvatarEmoji}
}

func toResponseVO(r TasteResponse) responseVO {
	photos := r.Photos
	if photos == nil {
		photos = []string{}
	}
	return responseVO{
		ID:        r.ID,
		WishID:    r.WishID,
		User:      toUserVO(r.User),
		Photos:    photos,
		Comment:   r.Comment,
		Rating:    r.Rating,
		IsAdopted: r.IsAdopted,
		CreatedAt: r.CreatedAt,
	}
}

func toWishVO(w Wish, responseCount int64, responses []TasteResponse) wishVO {
	vo := wishVO{
		ID:             w.ID,
		User:           toUserVO(w.User),
		RestaurantName: w.RestaurantName,
		Address:        w.Address,
		DishList:       w.DishList,
		Reason:         w.Reason,
		RewardCoins:    w.RewardCoins,
		Status:         w.Status,
		ResponseCount:  responseCount,
		CreatedAt:      w.CreatedAt,
	}
	if responses != nil {
		vo.Responses = make([]responseVO, 0, len(responses))
		for _, r := range responses {
			vo.Responses = append(vo.Responses, toResponseVO(r))
		}
	}
	return vo
}

// ---------- 通用响应 ----------

func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func failErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		fail(c, http.StatusNotFound, "记录不存在")
	case errors.Is(err, errInsufficient):
		fail(c, http.StatusBadRequest, "尝鲜币余额不足，先去赚一点吧")
	case errors.Is(err, errInvalidAmount):
		fail(c, http.StatusBadRequest, "金额必须大于 0")
	case errors.Is(err, errConflict):
		fail(c, http.StatusConflict, "该心愿当前状态不允许此操作")
	default:
		fail(c, http.StatusBadRequest, err.Error())
	}
}

// ---------- 认证 ----------

type registerReq struct {
	Username    string `json:"username" binding:"required,min=3,max=32"`
	Password    string `json:"password" binding:"required,min=6,max=64"`
	Nickname    string `json:"nickname" binding:"required,min=1,max=32"`
	AvatarEmoji string `json:"avatar_emoji"`
}

func handleRegister(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请填写完整的注册信息（用户名≥3位，密码≥6位）")
		return
	}
	var exists int64
	DB.Model(&User{}).Where("username = ?", req.Username).Count(&exists)
	if exists > 0 {
		fail(c, http.StatusConflict, "该用户名已被占用")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, "服务器开小差了")
		return
	}
	emoji := req.AvatarEmoji
	if emoji == "" {
		emoji = "🍜"
	}
	user, err := register(req.Username, string(hash), req.Nickname, emoji)
	if err != nil {
		fail(c, http.StatusInternalServerError, "注册失败，请稍后再试")
		return
	}
	token, err := generateToken(user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "登录态签发失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func handleLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请输入用户名和密码")
		return
	}
	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		fail(c, http.StatusUnauthorized, "用户名或密码不正确")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		fail(c, http.StatusUnauthorized, "用户名或密码不正确")
		return
	}
	token, err := generateToken(user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "登录态签发失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func handleMe(c *gin.Context) {
	var user User
	if err := DB.First(&user, currentUserID(c)).Error; err != nil {
		fail(c, http.StatusUnauthorized, "用户不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ---------- 心愿 ----------

func handleListWishes(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}

	query := DB.Model(&Wish{}).Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)

	var wishes []Wish
	if err := query.Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&wishes).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 批量统计每条心愿的回应数
	counts := map[uint]int64{}
	if len(wishes) > 0 {
		ids := make([]uint, 0, len(wishes))
		for _, w := range wishes {
			ids = append(ids, w.ID)
		}
		type row struct {
			WishID uint
			Cnt    int64
		}
		var rows []row
		DB.Model(&TasteResponse{}).
			Select("wish_id, COUNT(*) AS cnt").
			Where("wish_id IN ?", ids).
			Group("wish_id").
			Scan(&rows)
		for _, r := range rows {
			counts[r.WishID] = r.Cnt
		}
	}

	items := make([]wishVO, 0, len(wishes))
	for _, w := range wishes {
		items = append(items, toWishVO(w, counts[w.ID], nil))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "size": size})
}

func handleGetWish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var wish Wish
	if err := DB.Preload("User").First(&wish, id).Error; err != nil {
		fail(c, http.StatusNotFound, "心愿不存在")
		return
	}
	var responses []TasteResponse
	DB.Preload("User").
		Where("wish_id = ?", wish.ID).
		Order("is_adopted DESC, created_at DESC").
		Find(&responses)
	c.JSON(http.StatusOK, gin.H{"wish": toWishVO(wish, int64(len(responses)), responses)})
}

type createWishReq struct {
	RestaurantName string `json:"restaurant_name" binding:"required,max=64"`
	Address        string `json:"address" binding:"required,max=128"`
	DishList       string `json:"dish_list" binding:"required,max=255"`
	Reason         string `json:"reason" binding:"required,max=512"`
	RewardCoins    int    `json:"reward_coins" binding:"required"`
}

func handleCreateWish(c *gin.Context) {
	var req createWishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请完整填写餐厅、地址、心愿菜单、种草理由与悬赏币")
		return
	}
	wish, err := createWish(
		currentUserID(c),
		strings.TrimSpace(req.RestaurantName),
		strings.TrimSpace(req.Address),
		strings.TrimSpace(req.DishList),
		strings.TrimSpace(req.Reason),
		req.RewardCoins,
	)
	if err != nil {
		failErr(c, err)
		return
	}
	DB.Preload("User").First(wish, wish.ID)
	c.JSON(http.StatusOK, gin.H{"wish": toWishVO(*wish, 0, nil)})
}

type appendReq struct {
	Amount int `json:"amount" binding:"required"`
}

func handleAppendReward(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req appendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请输入追加的悬赏币数量")
		return
	}
	wish, err := appendReward(uint(id), currentUserID(c), req.Amount)
	if err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"wish": wish})
}

type createResponseReq struct {
	Photos  []string `json:"photos" binding:"required,min=1,max=6"`
	Comment string   `json:"comment" binding:"required,min=5,max=1024"`
	Rating  int      `json:"rating" binding:"required,min=1,max=5"`
}

func handleCreateResponse(c *gin.Context) {
	wishID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req createResponseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请上传至少 1 张现场实拍，并写下不少于 5 个字的口味点评")
		return
	}
	resp, err := submitResponse(uint(wishID), currentUserID(c), req.Photos, strings.TrimSpace(req.Comment), req.Rating)
	if err != nil {
		failErr(c, err)
		return
	}
	DB.Preload("User").First(resp, resp.ID)
	c.JSON(http.StatusOK, gin.H{"response": toResponseVO(*resp)})
}

type settleReq struct {
	ResponseID uint `json:"response_id" binding:"required"`
}

func handleSettle(c *gin.Context) {
	wishID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var req settleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请选择要采纳的代吃回应")
		return
	}
	if err := settleWish(uint(wishID), currentUserID(c), req.ResponseID); err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已采纳并结算，尝鲜币已送达代吃者"})
}

func handleCancelWish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := cancelWish(uint(id), currentUserID(c)); err != nil {
		failErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "心愿已取消，悬赏尝鲜币已退回"})
}

// ---------- 个人中心 ----------

func handleMyWishes(c *gin.Context) {
	uid := currentUserID(c)
	var wishes []Wish
	DB.Preload("User").Where("user_id = ?", uid).Order("created_at DESC").Find(&wishes)
	counts := map[uint]int64{}
	if len(wishes) > 0 {
		ids := make([]uint, 0, len(wishes))
		for _, w := range wishes {
			ids = append(ids, w.ID)
		}
		type row struct {
			WishID uint
			Cnt    int64
		}
		var rows []row
		DB.Model(&TasteResponse{}).
			Select("wish_id, COUNT(*) AS cnt").
			Where("wish_id IN ?", ids).
			Group("wish_id").
			Scan(&rows)
		for _, r := range rows {
			counts[r.WishID] = r.Cnt
		}
	}
	items := make([]wishVO, 0, len(wishes))
	for _, w := range wishes {
		items = append(items, toWishVO(w, counts[w.ID], nil))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleMyResponses(c *gin.Context) {
	uid := currentUserID(c)
	var responses []TasteResponse
	DB.Preload("User").Where("user_id = ?", uid).Order("created_at DESC").Find(&responses)

	// 关联心愿信息，便于展示"我替谁吃了什么"
	wishMap := map[uint]Wish{}
	if len(responses) > 0 {
		ids := make([]uint, 0, len(responses))
		for _, r := range responses {
			ids = append(ids, r.WishID)
		}
		var wishes []Wish
		DB.Where("id IN ?", ids).Find(&wishes)
		for _, w := range wishes {
			wishMap[w.ID] = w
		}
	}
	type myResponseVO struct {
		responseVO
		RestaurantName string `json:"restaurant_name"`
		RewardCoins    int    `json:"reward_coins"`
		WishStatus     string `json:"wish_status"`
	}
	items := make([]myResponseVO, 0, len(responses))
	for _, r := range responses {
		item := myResponseVO{responseVO: toResponseVO(r)}
		if w, ok := wishMap[r.WishID]; ok {
			item.RestaurantName = w.RestaurantName
			item.RewardCoins = w.RewardCoins
			item.WishStatus = w.Status
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleMyTransactions(c *gin.Context) {
	var txns []CoinTransaction
	DB.Where("user_id = ?", currentUserID(c)).
		Order("created_at DESC, id DESC").
		Limit(100).
		Find(&txns)
	c.JSON(http.StatusOK, gin.H{"items": txns})
}

// ---------- 文件上传 ----------

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func handleUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "请选择要上传的图片")
		return
	}
	if file.Size > 5<<20 {
		fail(c, http.StatusBadRequest, "图片不能超过 5MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		fail(c, http.StatusBadRequest, "仅支持 jpg / png / gif / webp 图片")
		return
	}
	name := uuid.NewString() + ext
	dst := filepath.Join(uploadDir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		fail(c, http.StatusInternalServerError, "图片保存失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}
