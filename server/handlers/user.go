package handlers

import (
	"net/http"
	"strconv"

	"taste-server/database"
	"taste-server/middleware"
	"taste-server/models"

	"github.com/gin-gonic/gin"
)

// MyCoinLogs 个人中心：尝鲜币流水
func MyCoinLogs(c *gin.Context) {
	uid := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var total int64
	q := database.DB.Model(&models.CoinTransaction{}).Where("user_id = ?", uid)
	q.Count(&total)

	var logs []models.CoinTransaction
	q.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&logs)

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "size": size, "list": logs})
}

// MyWishes 我发布的求代吃心愿
func MyWishes(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var wishes []models.Wish
	database.DB.Where("poster_id = ?", uid).Order("created_at DESC, id DESC").Find(&wishes)
	fillWishAggregates(wishes, uid)
	c.JSON(http.StatusOK, gin.H{"list": wishes})
}

// MyResponses 我替别人去吃的记录
func MyResponses(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var responses []models.TastingResponse
	database.DB.Where("taster_id = ?", uid).Order("created_at DESC, id DESC").Find(&responses)

	if len(responses) > 0 {
		wishIDs := map[uint]bool{}
		for _, r := range responses {
			wishIDs[r.WishID] = true
		}
		var wishes []models.Wish
		database.DB.Where("id IN ?", mapKeys(wishIDs)).Find(&wishes)
		wishMap := map[uint]models.Wish{}
		posterSet := map[uint]bool{}
		for _, w := range wishes {
			wishMap[w.ID] = w
			posterSet[w.PosterID] = true
		}
		var users []models.User
		database.DB.Where("id IN ?", mapKeys(posterSet)).Find(&users)
		userMap := map[uint]models.User{}
		for _, u := range users {
			userMap[u.ID] = u
		}
		for i := range responses {
			responses[i].Mine = true
			if w, ok := wishMap[responses[i].WishID]; ok {
				ww := w
				if u, ok := userMap[w.PosterID]; ok {
					uu := u
					ww.Poster = &uu
				}
				responses[i].Wish = &ww
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"list": responses})
}
