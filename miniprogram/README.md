# 附件通模板小程序

> 对应 V2.8 §15.2.4 + §15.2.8

## 项目结构

```
miniprogram/
├── app.js              # 入口：读取 ext_json 注入配置
├── app.json            # 全局配置（pages、tabBar、permission）
├── app.wxss            # 全局样式
├── ext.json.example    # ext_json 注入示例（platform 部署时覆盖）
├── project.config.json # 微信开发者工具配置
├── sitemap.json
├── pages/
│   ├── home/           # 作者品牌首页
│   ├── file-detail/    # ⭐ 文件详情页（核心：激励视频解锁下载）
│   └── about/          # 关于（隐私协议 / 客服 / UGC 声明入口）
├── utils/
│   ├── ad.js           # 广告 SDK 封装（开屏 / 激励视频 / 插屏）
│   └── api.js          # 后端 API 调用（apiBaseUrl 来自 ext_json）
└── preview/
    ├── privacy.md      # 隐私协议黄金版本
    └── ugc-declare.md  # UGC 内容声明黄金版本
```

## 开发与调试

1. 安装[微信开发者工具](https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html)
2. 复制 `ext.json.example` 为 `ext.json` 并填入测试 AppID + 测试广告位
3. 用开发者工具导入本目录
4. 修改 `project.config.json` 的 `appid` 为测试小程序 AppID

## 部署到作者小程序

由平台后端**自动**完成（V2.8 §15.2.8 自动化流水线）：

```
作者授权 → SetShareRatio + AgencyCreateAdunit + add_category
       → modifyserverdomain + setMpPrivacySetting
       → wxa/commit（注入定制化 ext_json）
       → wxa/submit_audit（带 ugc_declare 黄金版本）
       → 审核通过事件 → wxa/release → 上线
```

## ext_json 注入字段

平台后端 `service/wx_service.go::buildExtJSON()` 生成，覆盖：
- `extAppid` - 作者小程序 AppID
- `ext.platformUserId` - 作者在附件通的 user_id
- `ext.authorName` - 作者品牌名（同步到导航栏）
- `ext.primaryColor` - 主题色
- `ext.adUnitIds.{splash, rewardedVideo, interstitial}` - 三种广告位 ID
- `ext.apiBaseUrl` - 后端 API 域名
- `ext.customerServiceWxId` - 客服微信号
- `ext.privacyVersion` - 隐私协议版本

## 核心交互

### 文件详情页解锁流程

```
用户从公众号文章点击 → 跳转本小程序 file-detail 页
  ↓
loadFile() 拉文件元信息（含作者昵称）
  ↓
显示「观看广告解锁下载」按钮
  ↓
用户点击 → showRewardedVideo()
  ↓
完整观看 → setData({ unlocked: true })
  ↓
显示「下载附件」按钮
  ↓
wx.downloadFile + wx.openDocument 在小程序内预览/保存
```

### 广告位预埋

| 位置 | 时机 | 模板代码状态 |
|---|---|---|
| 开屏广告 | App.onLaunch 时 | 占位（V1.5 完整接入） |
| 激励视频 | 文件详情页解锁 | ✅ 已接入 |
| 插屏广告 | 文件预览结束 | 框架已就位（按需调用 ad.showInterstitial） |
