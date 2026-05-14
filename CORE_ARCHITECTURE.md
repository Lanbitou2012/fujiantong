# 附件通 — 项目核心控制文档（系统宪法 V3.0）

> **⚠️ 开发前置警告**
> 本文档是「附件通」项目最高指导原则。任何数据库设计、API 设计、UI 设计前必须先阅读本文档；业务规则变更必须先更新本文档再改代码。
> V3.0 相对 V2.8 的最大变化：**砍掉 V2.0~V2.8 八版叠加修订史、砍掉 P0 的「模板小程序生命周期」、入驻流程改为「扫码授权一跳」、12 张表压回 6 张**。被删内容并非否定，而是 V1.0 MVP 用不到，整体平移到 V1.5 P1。
> V2.0~V2.8 的全部历史与论证保留在 `CORE_ARCHITECTURE.v2.8.history.md`，老项目代码参考保留在 `e:\项目\小项目\公众号附件助手\`。

---

## 1. 项目定位

- **产品名称**：附件通
- **核心痛点**：微信公众号文章无法直接挂载附件（PDF/Word/Excel/压缩包等）。
- **产品形态**：PC Web 工作台（作者 + 平台管理员）+ 微信小程序（C 端读者下载 + B 端作者轻管理）。
- **核心愿景**：极简的"上传 → 复制 → 粘贴"体验圈住公众号作者；通过微信"流量主代运营"机制让 C 端下载流量自动变现，平台、作者、推广员三方共赢。

### 1.1 客群锁定（V2.1 已锁定）

**默认作者 = 个人运营者**：
- 个人订阅号运营者（占比约 70%，主力）
- 受雇于机构、用个人身份独享 72% 分成的运营者（占比约 25%）
- 大公司公关账号暂不主推（合规流程长）

附件通的小程序**统一以作者个人主体注册**（身份证 + 人脸 + 手机号实名）——绕开所有"企业代注册"死路。

---

## 2. 平台不可变事实（V2.0 立项已落地）

| 项 | 值 |
|---|---|
| 服务商主体 | 河南乐享数字生活服务有限公司 |
| 第三方平台 AppID | `wx037bc87a7245b35e` |
| 平台对外品牌 | 附件通 |
| 服务商域名 | `fujian.5g6g.top`（开发期）|
| 「流量主代运营」权限集 ID | **135** |
| 必须同时勾选的权限集 | **135**（流量主代运营）+ **17**（小程序代码管理）|
| 服务商结算账户邮箱 | `552002521@qq.com` |
| 平台向腾讯开票 | 增值税专用发票 → 腾讯科技（深圳）有限公司 |

---

## 3. 系统角色（**3 类**，不另建 promoters / admins 表）

```
C 端读者（独立人群，仅消费内容，无后台、不在 users 表）
        │  完全分离
        ▼
─────────────────────────────────────────
B 端用户（一张 users 表 + 叠加位）
  ├─ 基础角色：作者
  ├─ 叠加位 ①：is_promoter = true（推广员能力）
  └─ 叠加位 ②：is_admin    = true（平台管理员能力）
```

**三种 B 端身份**：

| 身份 | 字段组合 | 能力 |
|---|---|---|
| 普通作者 | `is_promoter=false, is_admin=false` | 上传附件、一键复制（文本超链接 / 卡片 / H5 三选一）、查看自己 72% 收益 |
| 作者 + 推广员 | `is_promoter=true, is_admin=false` | 上述 + 专属二维码 + 名下作者列表 + 推广佣金结算单 |
| 作者 + 管理员（主理人）| `is_admin=true` | 平台所有后台权限：作者管理、share_ratio 调整、推广员开关、违规处置、生成结算单 |

**主理人通常 `is_promoter=true && is_admin=true`**——既能自己发附件、又能拉作者、又能管平台。

### 3.1 推广员能力的获取（V1.0 仅人工开通）

V1.0 仅由主理人在后台 toggle `is_promoter` 标志位。**不做申请制、不做门槛解锁、不做 30 天冷却**——这些挪到 V1.5 P1。

---

## 4. 角色铁律（**5 条**，违反一条必重构）

1. **单层提成**：推广佣金仅依赖 `parent_promoter_user_id`，绝不递归向上。三级及以上为传销红线，不可谈判。
2. **CHECK + 递归防循环**：DDL 加 `CHECK (id <> parent_promoter_user_id)`；改 parent 时业务层向上 walk 防环。
3. **后端中间件实时鉴权**：`/api/promoter/*` `/api/admin/*` 必须每次 `SELECT users.is_promoter / is_admin` 实时校验，**不信任 JWT 快照**。
4. **主理人自营打标**：主理人名下若有作者，推广佣金照常计算，但结算行打 `is_self_promoter=true`，财务以"主理人个人劳务收入"出账。
5. **角色变更必入历史表**：`is_promoter` / `is_admin` 任何变更写 `user_role_history`，支撑 V1.5 按曝光日归属。

---

## 5. 商业模型（V2.0 核心）

### 5.1 资金路径（铁律）

```
广告主投放 100 → 微信扣 50% 渠道成本 → 50 元广告变现
        │
        └─ 按 share_ratio=28 自动拆账，腾讯直打两个账户：
           ├─ 28%（14 元）→ 平台对公账户（河南乐享）
           └─ 72%（36 元）→ 作者个人银行卡（已代扣代缴个税）
        ↓
  每半月（次月 1 / 16 号）邮件结算单
```

**关键事实**：
- 平台**不自建分账系统**——腾讯就是分账执行方。
- 平台与作者**互不接触资金、互不开票、互不打款**。
- `share_ratio=28` 必须在系统初始化时 `SetShareRatio(28)`，否则官方默认 0:10（平台拿 0 元，致命陷阱）。

### 5.2 推广员佣金（平台内分润，与腾讯无关）

- 推广员佣金 = 平台 28% 中切 30% = 作者收益总盘的 **8.4%**。
- **腾讯不参与推广员分账**——由平台从对公账户线下转账给推广员（金额小、对象散，合规风险低）。
- 系统职责：自动记账 + 月度生成结算单 + 主理人打款后标记 `settled=true`。
- **绝对不接微信支付分账接口**。

### 5.3 三类作者主体差异

| 主体 | 腾讯结算 | 是否需作者开票 | 文案 |
|---|---|---|---|
| 个人 | 直打银行卡，已代扣代缴个税 | 否 | "微信半月自动打款，已代扣代缴个税" |
| 个体工商户 | 走"委托代征代扣"通道 | 否 | "推荐委托代征代扣，操作最简" |
| 企业 | 作者开票后 30 工作日打款 | **是** | "需开增值税专票邮寄腾讯" |

**附件通统一引导个人主体**——这是产品策略基石。

---

## 6. 一键复制（V2.7 工程实证）

### 6.1 主链路：`weapp_text_link` 文本超链接

公众号 Web 编辑器识别带特定 class 的 `<a>` 标签，自动渲染为蓝色小程序文本超链接。**附件通标准结构**：

```html
<a class="weapp_text_link js_weapp_entry"
   data-miniprogram-type="text"
   data-miniprogram-appid="<作者小程序 appid>"
   data-miniprogram-path="pages/file-detail/file-detail?id=<文件ID>"
   data-miniprogram-nickname="附件通"
   data-miniprogram-servicetype="">岗位表2025.xlsx</a>
```

### 6.2 三级剪贴板回退（参考实现 `e:\项目\小项目\公众号附件助手\admin\src\views\author\AuthorFiles.vue:340-390`）

1. **优先**：`ClipboardItem API`（Chrome/Edge 现代版）—— 同时写 `text/html` + `text/plain`
2. **回退**：`contentEditable div + execCommand('copy')` —— 保留富文本格式
3. **兜底**：纯文本 `navigator.clipboard.writeText` / textarea

### 6.3 三种复制场景

| 按钮 | 生成内容 | 适用 |
|---|---|---|
| **📋 一键粘贴**（主）| `weapp_text_link` | 公众号文章正文 → 蓝色超链接 |
| **复制小程序卡片** | `<mp-miniprogram>` | 公众号自动回复 / 服务号推送 |
| **🔗 H5 备份** | `https://域名/f/<文件ID>` | 文章末尾备用 / 老版本微信 |

**作者无需在公众号关联小程序**——`weapp_text_link` 直达，跳出"非同主体最多关联 3 个"的限制。

---

## 7. 技术架构

| 层 | 选型 |
|---|---|
| 后端 | Go 1.21+ / Gin / GORM。单二进制 + PM2 守护 + 宝塔托管 |
| 前端 PC | Vue 3 + Vite + Element Plus + Pinia |
| 小程序 | 微信原生框架 |
| 数据库 | MySQL 8（单库，预留 `bound_appid` 租户字段）|
| 缓存 | Redis（热点下载链接防击穿；非必需，关闭后退化为直查 MySQL）|
| 存储 | 本地 / 腾讯云 COS 二选一，由 `STORAGE_DRIVER` 切换 |

### 7.1 端 → API 映射

```
C 端读者   ──→ 小程序  ──→ /api/v1/c/*       （无登录态，限频）
作者       ──→ PC Web  ──→ /api/v1/author/*  （JWT, role=author）
作者       ──→ 小程序  ──→ /api/v1/author/*  （JWT, role=author）
推广员     ──→ PC Web  ──→ /api/v1/promoter/*（JWT + 实时校验 is_promoter）
管理员     ──→ PC Web  ──→ /api/v1/admin/*   （JWT + 实时校验 is_admin）
微信平台   ──→ 服务端  ──→ /api/v1/wx/*      （消息加解密，独立鉴权）
```

---

## 8. 数据模型（**V1.0 共 6 张表**）

```
users
  id, wx_unionid, wx_openid_web, nickname, avatar_url, phone,
  is_promoter, is_admin,
  parent_promoter_user_id (单层、CHECK id<>self),
  bound_appid (作者授权的小程序 appid),
  authorizer_refresh_token (加密存储),
  admin_username, login_passphrase_hash (仅 is_admin=true 时使用),
  payment_info (推广员收款信息),
  status, created_at, updated_at, deleted_at

files
  id, user_id, name, ext, size_bytes, storage_key,
  download_code (短码), status (审核中/上架/下架/违规),
  media_check_trace_id, view_count, download_count,
  created_at

authorizations
  id, app_id, user_id, authorizer_refresh_token, status (authorized/deauthorized),
  granted_permission_ids, authorized_at, deauthorized_at, updated_at

wx_components                  -- 第三方平台单行配置 + token 中央存储
  id, key, access_token, refresh_token, expires_at, updated_at
  (key='component' / 'auth_<appid>' / 'verify_ticket')

settlement_records             -- 流量主收益流水（从 api_getsettlement 拉）
  id, app_id, user_id, settle_period (yyyy-mm-上/下),
  total_revenue_cents, platform_share_cents, author_share_cents,
  source (publisher_api/manual), created_at

commission_settlements         -- 推广员佣金月度结算单
  id, promoter_user_id, period (yyyy-mm),
  total_commission_cents (= Σ 名下作者收益 × 28% × 30%),
  status (pending/approved/paid),
  is_self_promoter, payment_proof, paid_at, created_at

user_role_history              -- 角色变更历史（铁律 5）
  id, user_id, field, old_value, new_value,
  changed_by_admin, reason, created_at
```

**已废弃 / 不要再创建**：`mp_template_versions`、`mp_deployments`、`mp_audits`、`daily_ad_stats`、`file_share_cards`、`promoter_applications`、`admin_action_log`、`fujian_user`、`access_logs`。这些 V1.5 P1 再开。

---

## 9. 核心业务闭环

1. **入驻**（**一跳完成**）：推广员发链接 `?p=<promoter_id>` → 落地页 → 作者点 CTA → 后端拼 `https://mp.weixin.qq.com/cgi-bin/componentloginpage?...&pre_auth_code=...&redirect_uri=...` → 作者跳微信扫码授权（**自动同时勾选 135 + 17 权限集**）→ 微信回跳 `/api/v1/wx/component/callback?auth_code=...` → 后端 `QueryAuth` 换 `authorizer_refresh_token` → 落库 user + 绑定 promoter → 签发 JWT → 跳前端 `/auth-success#token=...`。
2. **生产**：作者 PC 上传 → `mediaCheckAsync` 异步机审 → 复制 `weapp_text_link` HTML → 粘贴公众号正文。
3. **消费**：读者点超链接 → 小程序打开 → Redis 拿真实 URL（防击穿）→ 触发激励/插屏广告 → CDN 下载。
4. **协同**：作者小程序 `chooseMessageFile` 选新文件 → 替换（短码不变，链接不失效）。
5. **变现**：每半月 cron 拉 `api_getsettlement` 写 `settlement_records` → 月初生成 `commission_settlements` → 主理人审核 + 线下转账 → 标 `status=paid`。

---

## 10. 前端形态

### 10.1 H5 入驻：**1 页**（`/onboarding`）

落地页含 Hero + 价值主张 + 3 步流程 + CTA。CTA 直跳 `auth_url`，**不做 RegisterGuide / Authorize 中间页**。
没有小程序的作者：CTA 旁边一行小字 "没有小程序？前往 [mp.weixin.qq.com] 注册个人主体小程序（约 10 分钟）"，外链跳出。

### 10.2 admin / 工作台：路由表

| 路径 | 页面 | 角色 |
|---|---|---|
| `/login` | 登录（管理员账号密码 + 微信扫码登录） | 公开 |
| `/onboarding` | 作者入驻落地页 | 公开 |
| `/auth-success` | 微信授权回调落地页 | 公开 |
| `/f/:code` | C 端 H5 兜底下载页 | 公开 |
| `/dashboard` | 仪表盘（作者收益概览） | 所有 B 端 |
| `/files` | 我的附件（含一键复制三级剪贴板） | 所有 B 端 |
| `/my-revenue` | 我的收益 | 所有 B 端 |
| `/promoter/qrcode` `/promoter/authors` `/promoter/commission` | 推广员三页 | `is_promoter=true` |
| `/admin/hub` | **V2Hub 单页运维中心**（折叠 4 区：用户管理 / 结算审核 / 平台设置 / 第三方授权状态） | `is_admin=true` |

### 10.3 小程序：3 Tab

1. **首页**：品牌卡 + 轮播 + 最近浏览
2. **历史**：下载历史
3. **我的**：作者工作台入口（非作者隐藏）

---

## 11. 不做的事（V1.0 P0 边界，违反者 PR 标 `[P1]`）

- ❌ `fastregisterpersonalmp` 一键代注册（V1.5 评估）
- ❌ 模板小程序生命周期管理（`mp_deployments` / `mp_audits` / `mp_template_versions` 三表 + commit/submit_audit/release 流水线）→ V1.5
- ❌ `getAuthorizerList` / `api_getadposdetail` 自动同步 cron → V1.5
- ❌ 推广员申请制 + N/M 门槛 + 30 天冷却期 → V1.5
- ❌ 多语言 / 站内信 / IM / 评论
- ❌ 复杂权限系统（RBAC、菜单权限）—— 三角色叠加位足够
- ❌ 自建支付通道（一切结算走线下转账）
- ❌ 多层级佣金（合规红线）

---

## 12. V1.0 → V1.5 升级清单

V1.5 P1（在 V1.0 跑通 + 拿到 5-10 个真实作者后启动）：
1. 模板小程序生命周期（commit / submit_audit / release 全流水）
2. `getAuthorizerList` 每日 cron + `api_getadposdetail` 每日 cron + `api_getsettlement` 每半月 cron
3. 推广员申请制 + 门槛 + 冷却期 + `promoter_applications` 表
4. `fastregisterpersonalmp` 扫码代注册（评估后决定）
5. 推广收益按广告曝光日逐日归属（依赖 `user_role_history`）

---

## 13. 项目目录约定

- `_docs/` — 微信官方文档原文（`solution.txt` `guide.txt`），开发参考用，禁删
- `CORE_ARCHITECTURE.md` — 本文档（V3.0 当前生效版）
- `CORE_ARCHITECTURE.v2.8.history.md` — V2.0~V2.8 完整历史与论证，归档
- `server/` — Go 后端
- `admin/` — Vue 3 前端（同时承载 admin、author、promoter、onboarding 四类入口）

---

## 版本历史

- V1.0（已弃用，备份保留）：错误判断"自动分账不可行"
- V2.0：基于"流量主代运营"机制重写，确立 28:72 自动分账
- V2.1：客群锁定个人运营者；新增竞品分析与风险表
- V2.2：角色叠加位模型（取消独立 promoters 表）
- V2.3：8 条角色铁律
- V2.4：MVP 路线图（圣经页）
- V2.5：模板小程序全生命周期自动化（V3.0 已挪 P1）
- V2.6：类目锁定「工具→办公」+ 证据等级铁律
- V2.7：`weapp_text_link` 工程实证，复用上代项目
- V2.8：入驻路径锁定「作者自助注册 + 自动化授权」
- **V3.0（当前）**：移除 V2.0~V2.8 修订史堆叠（→ history 归档）、模板生命周期挪 P1、入驻流程改为"一跳完成"、12 表压回 6 表、铁律 8 → 5 条
