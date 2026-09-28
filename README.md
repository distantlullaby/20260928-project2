# 🍜 替你尝一口（Taste For You）

> 让忙碌的城市青年，先看「别人替你吃的现场」，再决定要不要亲自去。
> 通过他人的味蕾代尝心仪的餐厅与街角小吃，核心由「尝鲜币」循环生态驱动。

## ✨ 产品玩法

1. **注册即赠送 100 尝鲜币**（在同一数据库事务内到账并写流水）。
2. **发布「求代吃」**：填写餐厅、心愿菜单、种草理由与悬赏币，发布瞬间从可用余额
   **冻结**对应悬赏币。
3. **替 TA 去吃**：其他用户点击后上传**现场实拍 + 口味点评 + 星级**。
4. **采纳结算**：发起人翻看双面明信片背面的现场记录，确认采纳某位代吃者，
   系统把冻结的悬赏币结算给对方（同样在数据库事务内完成）。
5. **追加悬赏 / 取消退款**：等待中可随时追加悬赏（继续冻结），无人响应可取消，
   已冻结币全额退回。

首页 Feed 是可 **3D 翻转的双面明信片**：正面「心愿菜单 🌱 种草理由」，
翻面是「📷 现场实拍 + 口味点评」。

## 🧱 技术栈

- **前端**：Vue 3（`<script setup>`）+ Vite 5 + Vue Router + Axios
- **后端**：Go 1.22 + Gin + GORM（MySQL 8）
- **鉴权**：JWT（Bearer Token，bcrypt 存储密码）
- **资金安全**：所有硬币变动走 GORM 数据库事务 + `SELECT … FOR UPDATE` 行锁

## 📁 目录结构

```
.
├── backend/                Go + Gin + GORM 服务
│   ├── main.go             路由与启动
│   ├── config.go           配置（环境变量可覆盖）
│   ├── db.go               MySQL 连接 / AutoMigrate / 行锁子句
│   ├── models.go           4 张表：users / wishes / taste_responses / coin_transactions
│   ├── service.go          ★ 事务资金逻辑：注册赠送/冻结/追加/结算/退款
│   ├── handlers.go         HTTP 处理器与视图对象
│   ├── jwt.go              JWT 签发、解析与鉴权中间件
│   ├── seed.go             幂等种子数据（含 SVG 占位实拍）
│   └── smoke.ps1           34 项接口与资金账断言测试（PowerShell）
└── frontend/               Vue 3 + Vite
    └── src/
        ├── pages/          FeedPage（首页）/ LoginPage / ProfilePage（个人中心）
        ├── components/     WishCard（双面明信片）/ CreateWish / Taste / Append 弹层
        ├── api.js store.js router.js toast.js
        └── styles.css
```

## 🗄️ 数据模型（4 张表）

| 表 | 关键字段 | 说明 |
|---|---|---|
| `users` | `coins` 可用余额、`frozen_coins` 冻结额 | 注册送 100 |
| `wishes` | 心愿菜单、种草理由、`reward_coins`、`frozen_coins`、`status` | open / settled / canceled |
| `taste_responses` | `photos`(JSON 数组)、点评、星级、`is_adopted` | 代吃回应 |
| `coin_transactions` | 带符号 `amount`、`balance_after`、`type` | 完整硬币流水 |

**账目恒等式**：每个用户 `Σ coin_transactions.amount == users.coins`
（结算只划转冻结池，发起人侧记金额 0 的「结算完成」信息流水，不重复扣减可用余额）。

## 🔒 事务保障的资金动作

`backend/service.go` 中全部使用 `DB.Transaction(...)` 并对用户/心愿行加排他锁：

- `register`：建用户 + 100 币 + 赠送流水原子提交
- `createWish` / `appendReward`：锁用户行 → 校验余额 → 可用转冻结 → 写心愿 + 流水
- `settleWish`：锁心愿与双方用户行 → 校验状态 → 冻结池转入代吃者余额 →
  标记采纳/已结算 → 双方流水
- `cancelWish`：锁心愿与发起人 → 冻结退回可用 → 标记取消 + 退款流水

并发下不会超扣、不会重复结算，失败整体回滚。

## 🚀 本地运行

前置：Go 1.22、本地 MySQL 8（默认 `root/123456`）、Node 18+。

### 1. 数据库

MySQL 中创建库（服务首次启动也会自动建表）：

```sql
CREATE DATABASE taste_foryou DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 后端（:8080）

```bash
cd backend
go run .
# 或：go build -o app.exe . && ./app.exe
```

可用环境变量覆盖：`MYSQL_DSN`、`PORT`、`JWT_SECRET`、`UPLOAD_DIR`。
首次启动在空库自动写入种子数据。

### 3. 前端（:5173 / 被占用时顺延）

```bash
cd frontend
npm install
npm run dev
```

浏览器打开 Vite 提示的地址（已配置 `/api`、`/uploads` 代理到 8080）。

## 🎁 演示账号（密码均为 `123456`）

| 用户 | 角色 | 初始状态 |
|---|---|---|
| `ali` 爱逛吃的阿粒 | 发起人 | 发过 2 个心愿（1 已结算 / 1 已取消退款） |
| `bo` 夜宵雷达波波 | 代吃者 + 发起人 | 代吃被采纳赚过 30 币，自己也挂着 1 单 |
| `ci` 甜品捕手CiCi | 代吃者 | 提交过 1 份等待采纳的回应 |

## ✅ 接口与测试

- 接口烟雾/资金测试：`cd backend && powershell -ExecutionPolicy Bypass -File smoke.ps1`
  （34 项断言：注册赠送、冻结/追加/结算/退款、越权、重复结算、余额不足、
  流水恒等式等，需在空库运行）
- 公开：`POST /api/register`、`POST /api/login`、`GET /api/wishes`、`GET /api/wishes/:id`
- 鉴权：`GET /api/me`、`POST /api/wishes`、`/api/wishes/:id/append|cancel|settle`、
  `POST /api/wishes/:id/responses`、`GET /api/my/wishes|responses|transactions`、
  `POST /api/upload`（jpg/png/gif/webp，≤5MB）
