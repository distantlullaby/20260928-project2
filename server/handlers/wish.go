package handlers

import (
	"net/http"
	"strconv"

	"taste-server/database"
	"taste-server/middleware"
	"taste-server/models"
	"taste-server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createWishRequest struct {
	Restaurant string `json:"restaurant" binding:"required,max=128"`
	Dish       string `json:"dish" binding:"required,max=128"`
	Reason     string `json:"reason" binding:"required,min=1,max=1024"`
	Address    string `json:"address" binding:"max=256"`
	Bounty     int    `json:"bounty" binding:"min=1"`
}

// CreateWish 发布求代吃需求：默认冻结并扣除（从可用余额转入冻结）悬赏币
func CreateWish(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req createWishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请完整填写餐厅、心愿菜单、种草理由，且悬赏币至少为 1"})
		return
	}

	wish := models.Wish{
		PosterID:   uid,
		Restaurant: req.Restaurant,
		Dish:       req.Dish,
		Reason:     req.Reason,
		Address:    req.Address,
		Bounty:     req.Bounty,
		Status:     models.WishStatusOpen,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 先创建心愿拿到 ID，供流水关联
		if err := tx.Create(&wish).Error; err != nil {
			return err
		}
		// 锁定行 + 检查余额 + 冻结（可用余额 -n，冻结余额 +n）
		if _, err := services.WriteCoinTx(tx, uid, -req.Bounty, req.Bounty,
			models.CoinTypeFreeze, "wish", wish.ID, "发布代吃悬赏："+req.Restaurant); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == services.ErrInsufficient {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "尝鲜币余额不足，无法冻结悬赏"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wish": wish})
}

type addBountyRequest struct {
	Amount int `json:"amount" binding:"min=1"`
}

// AddBounty 追加悬赏：再次冻结
func AddBounty(c *gin.Context) {
	uid := middleware.GetUserID(c)
	wishID := paramUint(c, "id")
	var req addBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "追加的尝鲜币数量至少为 1"})
		return
	}

	var wish models.Wish
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&wish, wishID).Error; err != nil {
			return services.BizError("心愿不存在")
		}
		if wish.PosterID != uid {
			return services.BizError("只能给自己的心愿追加悬赏")
		}
		if wish.Status != models.WishStatusOpen {
			return services.BizError("心愿已结束，无法追加悬赏")
		}
		if _, err := services.WriteCoinTx(tx, uid, -req.Amount, req.Amount,
			models.CoinTypeFreeze, "wish", wish.ID, "追加代吃悬赏："+wish.Restaurant); err != nil {
			return err
		}
		return tx.Model(&wish).Update("bounty", gorm.Expr("bounty + ?", req.Amount)).Error
	})
	if err != nil {
		respondBiz(c, err, "追加悬赏失败")
		return
	}
	database.DB.First(&wish, wishID)
	c.JSON(http.StatusOK, gin.H{"wish": wish})
}

// CancelWish 取消心愿：冻结中的悬赏币全额退回
func CancelWish(c *gin.Context) {
	uid := middleware.GetUserID(c)
	wishID := paramUint(c, "id")

	var wish models.Wish
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&wish, wishID).Error; err != nil {
			return services.BizError("心愿不存在")
		}
		if wish.PosterID != uid {
			return services.BizError("只能取消自己的心愿")
		}
		if wish.Status != models.WishStatusOpen {
			return services.BizError("心愿已结束，无法取消")
		}
		var cnt int64
		tx.Model(&models.TastingResponse{}).Where("wish_id = ?", wishID).Count(&cnt)
		if cnt > 0 {
			return services.BizError("已有人替你去吃，请先采纳回应，不能直接取消")
		}
		// 冻结退回（冻结 -n，可用 +n）
		if _, err := services.WriteCoinTx(tx, uid, wish.Bounty, -wish.Bounty,
			models.CoinTypeUnfreeze, "wish", wish.ID, "取消心愿退回悬赏："+wish.Restaurant); err != nil {
			return err
		}
		return tx.Model(&wish).Updates(map[string]interface{}{"status": models.WishStatusCanceled}).Error
	})
	if err != nil {
		respondBiz(c, err, "取消失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AcceptWish 发起人确认采纳某条品尝回应：解冻扣款，把悬赏币结算给代吃者
func AcceptWish(c *gin.Context) {
	uid := middleware.GetUserID(c)
	wishID := paramUint(c, "id")
	respID := paramUint(c, "responseId")

	var wish models.Wish
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&wish, wishID).Error; err != nil {
			return services.BizError("心愿不存在")
		}
		if wish.PosterID != uid {
			return services.BizError("只有发起人能确认采纳")
		}
		if wish.Status != models.WishStatusOpen {
			return services.BizError("心愿已结束，不能重复采纳")
		}
		var resp models.TastingResponse
		if err := tx.First(&resp, respID).Error; err != nil {
			return services.BizError("品尝回应不存在")
		}
		if resp.WishID != wishID {
			return services.BizError("回应与心愿不匹配")
		}
		if resp.TasterID == uid {
			return services.BizError("不能采纳自己的回应")
		}

		// 1) 发起人：冻结余额扣除悬赏（frozen -n）
		if _, err := services.WriteCoinTx(tx, uid, 0, -wish.Bounty,
			models.CoinTypeSettleSpend, "wish", wish.ID,
			"采纳代吃，悬赏支出："+wish.Restaurant); err != nil {
			return err
		}
		// 2) 代吃者：可用余额增加悬赏
		if _, err := services.WriteCoinTx(tx, resp.TasterID, wish.Bounty, 0,
			models.CoinTypeSettleIncome, "wish", wishID,
			"代吃悬赏收入："+wish.Restaurant); err != nil {
			return err
		}
		// 3) 心愿标记已结算
		return tx.Model(&wish).Updates(map[string]interface{}{
			"status":               models.WishStatusSettled,
			"accepted_response_id": resp.ID,
		}).Error
	})
	if err != nil {
		respondBiz(c, err, "确认采纳失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListWishes Feed 流：求代吃需求列表（最新在前，支持 mine / status 过滤）
func ListWishes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}
	currentUID := middleware.GetUserID(c)

	q := database.DB.Model(&models.Wish{})
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if c.Query("mine") == "1" && currentUID != 0 {
		q = q.Where("poster_id = ?", currentUID)
	}

	var total int64
	q.Count(&total)

	var wishes []models.Wish
	q.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&wishes)

	fillWishAggregates(wishes, currentUID)

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"list":  wishes,
	})
}

// GetWish 心愿详情 + 全部品尝回应（明信片翻面用）
func GetWish(c *gin.Context) {
	currentUID := middleware.GetUserID(c)
	wishID := paramUint(c, "id")

	var wish models.Wish
	if err := database.DB.First(&wish, wishID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "心愿不存在"})
		return
	}
	wrapped := []models.Wish{wish}
	fillWishAggregates(wrapped, currentUID)
	wish = wrapped[0]

	var responses []models.TastingResponse
	database.DB.Where("wish_id = ?", wishID).Order("created_at DESC").Find(&responses)
	posterIDs := map[uint]bool{}
	for i := range responses {
		posterIDs[responses[i].TasterID] = true
		responses[i].Mine = currentUID != 0 && responses[i].TasterID == currentUID
	}
	var users []models.User
	database.DB.Where("id IN ?", mapKeys(posterIDs)).Find(&users)
	userMap := map[uint]models.User{}
	for _, u := range users {
		userMap[u.ID] = u
	}
	for i := range responses {
		if u, ok := userMap[responses[i].TasterID]; ok {
			uu := u
			responses[i].Taster = &uu
		}
	}

	c.JSON(http.StatusOK, gin.H{"wish": wish, "responses": responses})
}

// ---- 内部辅助 ----

func fillWishAggregates(wishes []models.Wish, currentUID uint) {
	if len(wishes) == 0 {
		return
	}
	ids := make([]uint, len(wishes))
	posterSet := map[uint]bool{}
	accIDs := []uint{}
	for i := range wishes {
		ids[i] = wishes[i].ID
		posterSet[wishes[i].PosterID] = true
		if wishes[i].AcceptedResponseID != nil {
			accIDs = append(accIDs, *wishes[i].AcceptedResponseID)
		}
	}

	var posters []models.User
	database.DB.Where("id IN ?", mapKeys(posterSet)).Find(&posters)
	posterMap := map[uint]models.User{}
	for _, u := range posters {
		posterMap[u.ID] = u
	}

	type cntRow struct {
		WishID uint
		Cnt    int
	}
	var cnts []cntRow
	database.DB.Model(&models.TastingResponse{}).
		Select("wish_id, COUNT(*) AS cnt").
		Where("wish_id IN ?", ids).Group("wish_id").Scan(&cnts)
	cntMap := map[uint]int{}
	for _, r := range cnts {
		cntMap[r.WishID] = r.Cnt
	}

	var accepted []models.TastingResponse
	if len(accIDs) > 0 {
		database.DB.Where("id IN ?", accIDs).Find(&accepted)
	}
	accMap := map[uint]models.TastingResponse{}
	tasterSet := map[uint]bool{}
	for _, r := range accepted {
		accMap[r.WishID] = r
		tasterSet[r.TasterID] = true
	}
	var tasters []models.User
	database.DB.Where("id IN ?", mapKeys(tasterSet)).Find(&tasters)
	tasterMap := map[uint]models.User{}
	for _, u := range tasters {
		tasterMap[u.ID] = u
	}

	for i := range wishes {
		if u, ok := posterMap[wishes[i].PosterID]; ok {
			uu := u
			wishes[i].Poster = &uu
		}
		wishes[i].RespCount = cntMap[wishes[i].ID]
		wishes[i].Mine = currentUID != 0 && wishes[i].PosterID == currentUID
		if r, ok := accMap[wishes[i].ID]; ok {
			rr := r
			if u, ok := tasterMap[r.TasterID]; ok {
				uu := u
				rr.Taster = &uu
			}
			wishes[i].Response = &rr
		}
	}
}

func paramUint(c *gin.Context, key string) uint {
	v, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return uint(v)
}

func mapKeys(m map[uint]bool) []uint {
	out := make([]uint, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// respondBiz 统一处理 service 层返回的错误
func respondBiz(c *gin.Context, err error, fallback string) {
	if biz, ok := services.IsBiz(err); ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": biz.Error()})
		return
	}
	if err == services.ErrInsufficient {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "尝鲜币余额不足"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}
