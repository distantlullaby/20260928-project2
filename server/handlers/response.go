package handlers

import (
	"net/http"

	"taste-server/database"
	"taste-server/middleware"
	"taste-server/models"
	"taste-server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createResponseRequest struct {
	Photos []string `json:"photos" binding:"required,min=1,dive,max=512"`
	Review string   `json:"review" binding:"required,min=1,max=2048"`
	Rating int      `json:"rating" binding:"min=1,max=5"`
}

// CreateResponse 替 TA 去吃：上传现场实拍与口味点评
func CreateResponse(c *gin.Context) {
	uid := middleware.GetUserID(c)
	wishID := paramUint(c, "id")

	var req createResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少上传一张现场实拍并填写口味点评"})
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		req.Rating = 5
	}

	resp := models.TastingResponse{
		WishID:   wishID,
		TasterID: uid,
		Photos:   req.Photos,
		Review:   req.Review,
		Rating:   req.Rating,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var wish models.Wish
		if err := tx.First(&wish, wishID).Error; err != nil {
			return services.BizError("心愿不存在")
		}
		if wish.Status != models.WishStatusOpen {
			return services.BizError("这个心愿已结束，不能再去代吃啦")
		}
		if wish.PosterID == uid {
			return services.BizError("不能替自己代吃哦")
		}
		var dup int64
		tx.Model(&models.TastingResponse{}).
			Where("wish_id = ? AND taster_id = ?", wishID, uid).Count(&dup)
		if dup > 0 {
			return services.BizError("你已经替 TA 吃过啦，可以在原回应上等待采纳")
		}
		return tx.Create(&resp).Error
	})
	if err != nil {
		respondBiz(c, err, "提交失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}
