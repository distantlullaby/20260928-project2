# 替你尝一口 🍜

> 帮助忙碌的城市青年，通过他人的味蕾代尝心仪的餐厅与街角小吃。
> 先看「别人替你吃的现场」，再决定要不要亲自去。

## 功能一览

### 两个 Web 页面

1. **首页 Feed 流**（`/feed`）
   - 展示「求代吃」需求列表，支持状态筛选（求代吃中 / 已采纳 / 已取消）与「只看我发的」
   - **双面明信片卡片**：正面是「心愿菜单 ＋ 种草理由」，心愿被采纳后可 3D 翻面查看「现场实拍 ＋ 口味点评」
   - 展开卡片可：替 TA 去吃、追加悬赏、取消退回、查看全部回应
2. **个人中心**（`/profile`）
   - 尝鲜币钱包：可用余额 + 冻结中余额
   - 尝鲜币流水（注册赠送 / 悬赏冻结 / 取消退回 / 采纳支出 / 代吃收入）
   - 我发布的心愿、我替吃的记录

### 尝鲜币循环生态

| 环节 | 账目变化（事务 + 行级锁） |
|---|---|
| 注册 | 赠送 **100** 枚尝鲜币，写入流水 |
| 发布代吃请求 | 可用余额 → 冻结（默认冻结悬赏，余额不足返回 402） |
| 追加悬赏 | 再次从可用余额冻结，支持多次追加 |
| 取消心愿（无人回应时） | 冻结币全额退回可用余额 |
| 代吃者提交回应 | 上传现场实拍（多图）+ 口味点评 + 星级 |
| 发起人确认采纳 | 同一事务内：发起人冻结扣款 → 代吃者入账 + 心愿标记 settled |

## 技术栈

- **前端**：Vue 3（Composition API）+ Vite + Vue Router + Axios
- **后端**：Go 1.22 + Gin + GORM（MySQL 8）
- **认证**：JWT（Bearer Token）
- **图片**：本地上传 `POST /api/upload`，静态服务 `/uploads`

## 数据表

- `users` 用户表（含 `coin_balance` 可用 / `frozen_balance` 冻结两个余额）
- `wishes` 心愿表（悬赏额、状态、采纳回应 ID）
- `tasting_responses` 品尝回应表（实拍 JSON 数组、点评、打分）
- `coin_transactions` 尝鲜币流水表（金额、变动后双余额快照、类型、关联单据）

## 目录结构

```
├── server/                 Go + Gin + GORM
│   ├── main.go             路由入口
│   ├── config/             配置（DSN/端口/注册赠送额）
│   ├── database/           连接与自动建表
│   ├── models/             四张表模型
│   ├── middleware/         JWT 必选 / 可选鉴权
│   ├── services/           尝鲜币事务核心（行级锁 + 流水）
│   ├── handlers/           auth / wish / response / user / upload
│   └── e2e_test.mjs        端到端接口测试（30 个断言）
└── web/                    Vue 3 + Vite
    └── src/
        ├── api/            Axios 封装与接口
        ├── store/          auth / toast / ui 全局状态
        ├── router/         路由与登录守卫
        ├── components/     双面明信片、各业务弹窗
        └── views/          FeedView / ProfileView
```

## 本地启动

### 1. 数据库

MySQL 已在本地启动（默认 `root/123456`），创建数据库（服务首次启动也会自动建表）：

```sql
CREATE DATABASE taste_for_you DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

配置可用环境变量覆盖：`MYSQL_DSN`、`JWT_SECRET`、`LISTEN_ADDR`、`UPLOAD_DIR`。

### 2. 启动后端（:8081）

```powershell
# Go 位于 C:\mine\code\tool\Google\go\go1.22.1\sdk\bin
cd server
go run .
```

### 3. 启动前端（:5180）

```powershell
# Node 由 nvm 管理
cd web
npm install
npm run dev
```

打开 http://127.0.0.1:5180 ，注册两个账号即可体验完整循环：

1. A 注册领 100 币 → 发布心愿冻结悬赏
2. B 注册领 100 币 → 在 Feed 点「替 TA 去吃」上传实拍与点评
3. A 在「全部回应」里点确认采纳 → 悬赏币结算给 B
4. 双方到「个人中心」查看余额与流水

### 4. 接口自测

后端启动后：

```powershell
cd server
node e2e_test.mjs
```

覆盖注册赠币、冻结、余额不足拒绝、追加、代吃回应、采纳结算、取消退回、防重复采纳、流水与个人记录等 30 个断言。

## RESTful API 摘要

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/api/register` | 注册（赠 100 币） | - |
| POST | `/api/login` | 登录 | - |
| GET | `/api/me` | 当前用户与余额 | ✓ |
| GET | `/api/wishes` | Feed 列表（`status` / `mine` / 分页） | 可选 |
| GET | `/api/wishes/:id` | 心愿详情 + 全部回应 | 可选 |
| POST | `/api/wishes` | 发布心愿（冻结悬赏） | ✓ |
| POST | `/api/wishes/:id/bounty` | 追加悬赏 | ✓ |
| POST | `/api/wishes/:id/cancel` | 取消并退回 | ✓ |
| POST | `/api/wishes/:id/responses` | 替 TA 去吃 | ✓ |
| POST | `/api/wishes/:id/accept/:responseId` | 确认采纳并结算 | ✓ |
| POST | `/api/upload` | 上传现场实拍 | ✓ |
| GET | `/api/my/coins` | 尝鲜币流水 | ✓ |
| GET | `/api/my/wishes` | 我发布的心愿 | ✓ |
| GET | `/api/my/responses` | 我替吃的记录 | ✓ |
