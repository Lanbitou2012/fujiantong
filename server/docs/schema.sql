-- ==========================================================
-- 附件通 (FujianTong) 核心数据库表结构 V2.8
-- 引擎: InnoDB | 字符集: utf8mb4
-- 对标: CORE_ARCHITECTURE.md §15.2.1
-- ==========================================================

-- 1. B 端用户统一表
CREATE TABLE `users` (
  `id`                        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `wx_union_id`               VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '微信 UnionID',
  `wx_open_id_web`            VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '微信网页端 OpenID',
  `nickname`                  VARCHAR(100) NOT NULL DEFAULT '',
  `avatar_url`                VARCHAR(512) NOT NULL DEFAULT '',
  `phone`                     VARCHAR(20)  NOT NULL DEFAULT '',

  `is_promoter`               TINYINT(1) NOT NULL DEFAULT 0 COMMENT '叠加权限：推广员',
  `is_admin`                  TINYINT(1) NOT NULL DEFAULT 0 COMMENT '叠加权限：管理员',

  `parent_promoter_user_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定推广员（铁律 2：单层）',

  `bound_appid`               VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '绑定的小程序 AppID',
  `authorizer_refresh_token`  VARCHAR(512) NOT NULL DEFAULT '' COMMENT '授权 refresh_token',

  `admin_username`            VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '管理员登录名',
  `login_passphrase_hash`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录密码 bcrypt hash',

  `payment_info`              TEXT         COMMENT '收款信息（JSON）',

  `status`                    TINYINT NOT NULL DEFAULT 1 COMMENT '1=正常 0=封禁',
  `created_at`                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`                TIMESTAMP NULL DEFAULT NULL,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_union_id` (`wx_union_id`),
  UNIQUE KEY `uk_bound_appid` (`bound_appid`),
  KEY `idx_promoter` (`parent_promoter_user_id`),
  KEY `idx_is_promoter` (`is_promoter`),
  KEY `idx_is_admin` (`is_admin`),
  KEY `idx_admin_username` (`admin_username`),
  KEY `idx_deleted_at` (`deleted_at`),
  -- 铁律 3：防止自引用
  CONSTRAINT `chk_no_self_ref` CHECK (`id` <> `parent_promoter_user_id` OR `parent_promoter_user_id` = 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='B 端用户统一表（作者 + 推广员叠加 + 管理员叠加）';


-- 2. 角色位变更历史（铁律 7）
CREATE TABLE `user_role_history` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `field`            VARCHAR(32)  NOT NULL COMMENT 'is_promoter / is_admin',
  `old_value`        VARCHAR(16)  NOT NULL DEFAULT '',
  `new_value`        VARCHAR(16)  NOT NULL DEFAULT '',
  `changed_by_admin` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `reason`           VARCHAR(512) NOT NULL DEFAULT '',
  `created_at`       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_field_time` (`user_id`, `field`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色位变更历史';


-- 3. 附件元信息表
CREATE TABLE `files` (
  `id`                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `author_user_id`       BIGINT UNSIGNED NOT NULL,
  `appid`                VARCHAR(64) NOT NULL DEFAULT '' COMMENT '租户字段',
  `name`                 VARCHAR(255) NOT NULL,
  `original_name`        VARCHAR(255) NOT NULL,
  `ext`                  VARCHAR(20)  NOT NULL,
  `size`                 BIGINT NOT NULL DEFAULT 0,
  `mime_type`            VARCHAR(120) NOT NULL DEFAULT '',
  `storage_type`         VARCHAR(32)  NOT NULL DEFAULT 'local',
  `storage_path`         VARCHAR(512) NOT NULL,
  `download_code`        VARCHAR(32)  NOT NULL,
  `view_count`           BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `download_count`       BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `media_check_status`   VARCHAR(32)  NOT NULL DEFAULT 'pending',
  `media_check_trace_id` VARCHAR(128) NOT NULL DEFAULT '',
  `status`               VARCHAR(32)  NOT NULL DEFAULT 'active',
  `created_at`           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`           TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_download_code` (`download_code`),
  KEY `idx_author` (`author_user_id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='附件元信息表';


-- 4. 分享卡片快照
CREATE TABLE `file_share_cards` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `file_id`    BIGINT UNSIGNED NOT NULL,
  `card_type`  VARCHAR(32)  NOT NULL COMMENT 'weapp_text_link / mp_miniprogram / h5_backup / miniprogram_path',
  `html`       LONGTEXT     NOT NULL,
  `appid`      VARCHAR(64)  NOT NULL DEFAULT '',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_file` (`file_id`),
  KEY `idx_appid` (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='复制内容快照（文本超链接 / 小程序卡片 / H5 备份）';


-- 5. 作者小程序授权记录
CREATE TABLE `authorizations` (
  `id`                        BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`                   BIGINT UNSIGNED NOT NULL,
  `appid`                     VARCHAR(64)  NOT NULL,
  `authorizer_refresh_token`  VARCHAR(512) NOT NULL DEFAULT '',
  `share_ratio`               INT NOT NULL DEFAULT 28,
  `share_ratio_confirmed_at`  TIMESTAMP NULL DEFAULT NULL,
  `mp_nickname`               VARCHAR(100) NOT NULL DEFAULT '',
  `mp_head_img`               VARCHAR(512) NOT NULL DEFAULT '',
  `principal_name`            VARCHAR(100) NOT NULL DEFAULT '',
  `publisher_status`          VARCHAR(32)  NOT NULL DEFAULT 'inactive',
  `granted_permission_ids`    VARCHAR(255) NOT NULL DEFAULT '',
  `status`                    VARCHAR(32)  NOT NULL DEFAULT 'authorized',
  `authorized_at`             TIMESTAMP NULL DEFAULT NULL,
  `deauthorized_at`           TIMESTAMP NULL DEFAULT NULL,
  `created_at`                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_appid` (`appid`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作者小程序授权记录';


-- 6. 每日广告数据
CREATE TABLE `daily_ad_stats` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `appid`           VARCHAR(64)  NOT NULL,
  `date`            VARCHAR(10)  NOT NULL COMMENT 'YYYY-MM-DD',
  `slot_id`         VARCHAR(64)  NOT NULL DEFAULT '',
  `ad_slot`         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'rewarded_video / interstitial / splash',
  `exposure_count`  BIGINT NOT NULL DEFAULT 0,
  `click_count`     BIGINT NOT NULL DEFAULT 0,
  `income`          BIGINT NOT NULL DEFAULT 0 COMMENT '单位：分',
  `ecpm`            DECIMAL(10,4) NOT NULL DEFAULT 0,
  `created_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_appid_date` (`appid`, `date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每日广告数据（api_getadposdetail）';


-- 7. 结算记录
CREATE TABLE `settlement_records` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `appid`           VARCHAR(64) NOT NULL,
  `settle_period`   VARCHAR(20) NOT NULL,
  `total_revenue`   BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `platform_share`  BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `author_share`    BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `user_id`         BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `raw_json`        LONGTEXT,
  `created_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_period` (`settle_period`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='结算记录（api_getsettlement）';


-- 8. 推广员佣金结算单
CREATE TABLE `commission_settlements` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `promoter_id`           BIGINT UNSIGNED NOT NULL,
  `settle_period`         VARCHAR(20) NOT NULL,
  `total_author_revenue`  BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `platform_fee`          BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `commission`            BIGINT NOT NULL DEFAULT 0 COMMENT '分',
  `is_self_promoter`      TINYINT(1) NOT NULL DEFAULT 0 COMMENT '铁律 5：主理人自营标记',
  `status`                VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/approved/paid/rejected',
  `transfer_proof`        VARCHAR(512) NOT NULL DEFAULT '',
  `paid_at`               TIMESTAMP NULL DEFAULT NULL,
  `approved_by`           BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `detail_json`           LONGTEXT,
  `created_at`            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_promoter` (`promoter_id`),
  KEY `idx_period` (`settle_period`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='推广员佣金结算单';


-- 9. 微信 Token 中央存储
CREATE TABLE `wx_tokens` (
  `key`           VARCHAR(64) NOT NULL COMMENT 'component 或 作者 appid',
  `access_token`  VARCHAR(512) NOT NULL DEFAULT '',
  `refresh_token` VARCHAR(512) NOT NULL DEFAULT '',
  `expires_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `extra`         TEXT COMMENT 'verify_ticket 等',
  `updated_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='微信 Token 中央存储';


-- 10. 模板小程序代码版本管理
CREATE TABLE `mp_template_versions` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_id`   BIGINT NOT NULL COMMENT '微信 template_id',
  `user_version`  VARCHAR(64)  NOT NULL,
  `user_desc`     VARCHAR(512) NOT NULL DEFAULT '',
  `status`        VARCHAR(32)  NOT NULL DEFAULT 'draft' COMMENT 'draft / active / deprecated',
  `change_log`    TEXT,
  `created_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模板小程序代码版本管理';


-- 11. 作者小程序部署记录
CREATE TABLE `mp_deployments` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`               BIGINT UNSIGNED NOT NULL,
  `appid`                 VARCHAR(64) NOT NULL,
  `template_version_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `template_id`           BIGINT NOT NULL DEFAULT 0,
  `user_version`          VARCHAR(64) NOT NULL DEFAULT '',
  `ext_json`              LONGTEXT,
  `status`                VARCHAR(64) NOT NULL DEFAULT 'pending',
  `config_result`         TEXT,
  `audit_id`              BIGINT NOT NULL DEFAULT 0,
  `fail_reason`           TEXT,
  `event_log`             LONGTEXT,
  `created_at`            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作者小程序部署记录';


-- 12. 提审记录
CREATE TABLE `mp_audits` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `appid`           VARCHAR(64) NOT NULL,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `deployment_id`   BIGINT UNSIGNED NOT NULL,
  `wx_audit_id`     BIGINT NOT NULL DEFAULT 0,
  `version_desc`    VARCHAR(512) NOT NULL DEFAULT '',
  `preview_info`    TEXT,
  `ugc_declare`     TEXT,
  `status`          VARCHAR(32) NOT NULL DEFAULT 'submitted' COMMENT 'submitted/success/fail/delay/undone',
  `fail_reason`     TEXT,
  `fail_category`   VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'code / qualification / unknown',
  `retry_count`     INT NOT NULL DEFAULT 0,
  `submitted_at`    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `result_at`       TIMESTAMP NULL DEFAULT NULL,
  `created_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_appid` (`appid`),
  KEY `idx_user` (`user_id`),
  KEY `idx_deployment` (`deployment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提审记录';


-- 13. 主理人操作审计日志
CREATE TABLE `admin_action_log` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_id`   BIGINT UNSIGNED NOT NULL,
  `action`     VARCHAR(64) NOT NULL,
  `target`     VARCHAR(128) NOT NULL DEFAULT '',
  `detail`     LONGTEXT,
  `ip`         VARCHAR(64) NOT NULL DEFAULT '',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_admin` (`admin_id`),
  KEY `idx_action` (`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主理人操作审计日志';
