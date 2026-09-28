package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// seedSVG 生成一张"现场实拍"占位图（暖色渐变 + 食物 emoji）。
func seedSVG(name, emoji, c1, c2 string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="450" viewBox="0 0 600 450">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%%" stop-color="%s"/>
      <stop offset="100%%" stop-color="%s"/>
    </linearGradient>
  </defs>
  <rect width="600" height="450" rx="24" fill="url(#g)"/>
  <text x="300" y="225" font-size="150" text-anchor="middle" dominant-baseline="central">%s</text>
  <rect x="380" y="24" width="196" height="44" rx="22" fill="rgba(0,0,0,0.35)"/>
  <text x="478" y="52" font-size="22" text-anchor="middle" fill="#fff" font-family="sans-serif">📷 现场实拍</text>
  <text x="300" y="410" font-size="20" text-anchor="middle" fill="rgba(255,255,255,0.85)" font-family="sans-serif">%s</text>
</svg>`, c1, c2, emoji, name)
}

// seed 仅在空库时执行，保证幂等。所有资金变动走事务服务函数，账目天然一致。
func seed() {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	seedDir := filepath.Join(uploadDir, "seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		log.Printf("种子图片目录创建失败: %v", err)
	}
	writePhoto := func(file, name, emoji, c1, c2 string) string {
		path := filepath.Join(seedDir, file)
		if err := os.WriteFile(path, []byte(seedSVG(name, emoji, c1, c2)), 0o644); err != nil {
			log.Printf("种子图片写入失败 %s: %v", file, err)
		}
		return "/uploads/seed/" + file
	}
	photo := map[string][]string{
		"noodle": {writePhoto("noodle.svg", "深夜小面", "🍜", "#ff9a6b", "#e2573d")},
		"bbq":    {writePhoto("bbq.svg", "路口烧烤摊", "🍢", "#ffb15c", "#d96b2a")},
		"cake":   {writePhoto("cake.svg", "海盐蛋糕", "🍰", "#ffc4d6", "#e8789c")},
		"burger": {writePhoto("burger.svg", "厚牛堡", "🍔", "#ffce7a", "#d98a2b")},
	}

	mkUser := func(username, nickname, emoji string) *User {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		u, err := register(username, string(hash), nickname, emoji)
		if err != nil {
			log.Fatalf("种子用户创建失败 %s: %v", username, err)
		}
		return u
	}
	ali := mkUser("ali", "爱逛吃的阿粒", "🍜")
	bo := mkUser("bo", "夜宵雷达波波", "🍢")
	ci := mkUser("ci", "甜品捕手CiCi", "🍰")

	mkWish := func(u *User, restaurant, address, dishes, reason string, reward int) *Wish {
		w, err := createWish(u.ID, restaurant, address, dishes, reason, reward)
		if err != nil {
			log.Fatalf("种子心愿创建失败: %v", err)
		}
		// 让示例数据看起来像是陆续发布的
		DB.Model(w).Update("created_at", time.Now().Add(-time.Duration(reward)*time.Hour))
		return w
	}

	w1 := mkWish(ali, "巷子里小面", "朝阳区幸福三村北街 12 号",
		"豌杂面、红油抄手、冰粉", "刷到无数次豌杂面的浇头拌面视频，据说麻辣鲜香、豌泥沙糯，加班完就馋这一口，求帮看真实分量和辣度！",
		30)
	w2 := mkWish(bo, "十字路口烧烤摊", "海淀区五道口地铁 B 口外 50 米",
		"烤五花、烤茄子、烤韭菜", "路过总在排队的露天烧烤，烟火气拉满，想知道肉串新鲜不新鲜、酱料是偏甜还是偏咸。",
		20)
	w3 := mkWish(ci, "一口甜蛋糕房", "西城区什刹海烟袋斜街 8 号",
		"海盐奥利奥蛋糕、巴斯克", "外卖图拍得很治愈，但怕奶油太腻、蛋糕体偏干，想请附近的姐妹替我先尝一块。",
		50)
	w4 := mkWish(ali, "周末厚牛堡", "东城区东直门来福士 B1",
		"双层和牛堡、薯角", "新品宣传说肉饼厚到盖不住面包，好奇汁水和性价比，求一个真实买家秀。",
		15)

	// 已有人替吃的回应
	addResp := func(wishID, tasterID uint, photos []string, comment string, rating int) *TasteResponse {
		r, err := submitResponse(wishID, tasterID, photos, comment, rating)
		if err != nil {
			log.Fatalf("种子回应创建失败: %v", err)
		}
		return r
	}
	r1 := addResp(w1.ID, bo.ID, photo["noodle"],
		"替你吃了！豌杂面分量很足，豌杂炖得沙沙的裹满面条，红油香而不呛、麻味偏重，不能吃辣建议微辣。抄手皮薄馅嫩，冰粉甜度刚好解辣，值得专门跑一趟。", 5)
	addResp(w2.ID, ci.ID, photo["bbq"],
		"现场烟火气确实足！五花肉烤得焦香不腻，酱料是咸鲜挂的、带一点点甜；茄子蒜蓉给得足。美中不足是等了 20 分钟，建议错峰。", 4)
	addResp(w3.ID, bo.ID, photo["cake"],
		"海盐奥利奥比想象中清爽，奶油是咸甜口不会腻，蛋糕体湿润不干；巴斯克焦香浓郁、偏甜一点点，配美式刚好。门店只有两三个座位，建议外带。", 5)

	// w1：发起人采纳波波的回应并结算（悬赏 30 币到账波波）
	if err := settleWish(w1.ID, ali.ID, r1.ID); err != nil {
		log.Fatalf("种子结算失败: %v", err)
	}
	// w4：取消并退回，演示 canceled 状态
	if err := cancelWish(w4.ID, ali.ID); err != nil {
		log.Fatalf("种子取消失败: %v", err)
	}

	log.Println("种子数据已生成（演示账号：ali / bo / ci，密码均为 123456）")
}
