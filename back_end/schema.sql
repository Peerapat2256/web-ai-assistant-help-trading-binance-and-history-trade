-- ====================================================
-- Binance Trading Journal & AI Analytics System
-- Database Schema for TiDB Cloud / MySQL
-- ====================================================

CREATE DATABASE IF NOT EXISTS binance_journal CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE binance_journal;

-- 1. ตารางผู้ใช้งาน (Member)
CREATE TABLE IF NOT EXISTS `member` (
    `member_id` INT AUTO_INCREMENT PRIMARY KEY,
    `fullname` VARCHAR(150) NOT NULL,
    `email` VARCHAR(191) NOT NULL UNIQUE,
    `password` VARCHAR(255) DEFAULT '',
    `phone_number` VARCHAR(50) DEFAULT '',
    `google_id` VARCHAR(191) DEFAULT '',
    `is_password_set` BOOLEAN DEFAULT TRUE,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_email` (`email`),
    INDEX `idx_google_id` (`google_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. ตารางการตั้งค่าพอร์ตและเป้าหมาย (Portfolio Setting)
CREATE TABLE IF NOT EXISTS `portfolio_setting` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `member_id` INT NOT NULL UNIQUE,
    `initial_balance` DECIMAL(18, 4) DEFAULT 1000.0000,
    `currency` VARCHAR(20) DEFAULT 'USDT',
    `monthly_target_pct` DECIMAL(6, 2) DEFAULT 10.00,
    `max_daily_loss_usd` DECIMAL(18, 4) DEFAULT 50.0000,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT `fk_portfolio_member` FOREIGN KEY (`member_id`) REFERENCES `member` (`member_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. ตารางประวัติการเทรด (Trade History)
CREATE TABLE IF NOT EXISTS `trade_history` (
    `trade_id` INT AUTO_INCREMENT PRIMARY KEY,
    `trade_code` VARCHAR(50) NOT NULL UNIQUE, -- เช่น TRD-202611-0001
    `member_id` INT NOT NULL,
    `symbol` VARCHAR(30) NOT NULL,            -- เช่น BTCUSDT, ETHUSDT
    `position_side` VARCHAR(10) NOT NULL,     -- LONG / SHORT / SPOT
    `leverage` INT DEFAULT 1,                 -- 1x - 125x
    `entry_price` DECIMAL(18, 6) NOT NULL,
    `exit_price` DECIMAL(18, 6) NOT NULL,
    `quantity` DECIMAL(18, 6) DEFAULT 0,
    `margin_used` DECIMAL(18, 4) DEFAULT 0,
    `pnl` DECIMAL(18, 4) NOT NULL,            -- กำไร/ขาดทุนสุทธิ (USDT)
    `pnl_percentage` DECIMAL(10, 4) NOT NULL, -- ROI %
    `fee` DECIMAL(18, 4) DEFAULT 0,
    `trade_time` DATETIME NOT NULL,
    `setup_tag` VARCHAR(100) DEFAULT '',      -- Breakout, Support/Resistance, SMC, etc.
    `emotion_tag` VARCHAR(100) DEFAULT '',    -- Follow Plan, FOMO, Revenge, etc.
    `notes` TEXT,
    `screenshot_url` VARCHAR(255) DEFAULT '', -- เก็บ path รูปแคปจอ Binance
    `sync_sheet_status` VARCHAR(20) DEFAULT 'Synced', -- Synced / Pending / Failed
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_trade_member` (`member_id`),
    INDEX `idx_trade_symbol` (`symbol`),
    INDEX `idx_trade_time` (`trade_time`),
    CONSTRAINT `fk_trade_member` FOREIGN KEY (`member_id`) REFERENCES `member` (`member_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. ตารางบันทึกการวิเคราะห์ของ AI (AI Analysis Log)
CREATE TABLE IF NOT EXISTS `ai_analysis_log` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `member_id` INT NOT NULL,
    `symbol` VARCHAR(30) NOT NULL,
    `timeframe` VARCHAR(20) DEFAULT '1h',
    `analysis_type` VARCHAR(50) NOT NULL,     -- Chart / News / TradeJournalReview
    `prompt_summary` TEXT,
    `ai_response` LONGTEXT NOT NULL,
    `signal_direction` VARCHAR(20) DEFAULT 'NEUTRAL', -- BULLISH / BEARISH / NEUTRAL
    `sentiment_score` DECIMAL(4, 2) DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_ai_member` (`member_id`),
    CONSTRAINT `fk_ai_member` FOREIGN KEY (`member_id`) REFERENCES `member` (`member_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. ตารางการตั้งค่าเชื่อมต่อ Google Sheets (Sheet Config)
CREATE TABLE IF NOT EXISTS `sheet_config` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `member_id` INT NOT NULL UNIQUE,
    `spreadsheet_id` VARCHAR(191) NOT NULL,
    `sheet_name` VARCHAR(100) DEFAULT 'Trades',
    `is_auto_sync` BOOLEAN DEFAULT TRUE,
    `last_synced_at` DATETIME NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT `fk_sheet_member` FOREIGN KEY (`member_id`) REFERENCES `member` (`member_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
