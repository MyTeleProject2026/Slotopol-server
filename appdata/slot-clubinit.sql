-- ============================================================
-- Slotopol Database Initialization Script
-- ============================================================

/*
Tables of databases creates in code if them absent.
Statements below with inserting executes if `club` table is empty.
Users access levels (`gal` or `access` fields) are sum of followed ints:
   1 - *member*, user have access to club
   2 - *dealer*, can change club game settings
   4 - *booker*, can change user balance and move user money to/from club deposit
   8 - *master*, can change club bank, fund, deposit
  16 - *admin*, can change same access levels to other users
  31 - all rights
*/

-- Insert default club
INSERT INTO `club` (`cid`, `name`, `bank`, `fund`, `lock`, `rate`, `mrtp`) VALUES
(1, 'virtual', 1000000, 100000, 0, 2.5, 95),
(2, 'global', 0, 0, 0, 2.5, 0);

-- Insert default users
INSERT INTO `user` (`uid`, `email`, `secret`, `name`, `status`, `gal`) VALUES
(1, 'admin@slotopol.com', 'admin123', 'Administrator', 1, 31),
(2, 'dealer@slotopol.com', 'dealer123', 'Dealer', 1, 3),
(3, 'player@slotopol.com', 'player123', 'Player', 1, 1);

-- Insert user properties
INSERT INTO `props` (`cid`, `uid`, `wallet`, `access`, `mrtp`) VALUES
(1, 1, 100000, 31, 0),
(2, 1, 0, 0, 0),
(1, 2, 10000, 13, 0),
(2, 2, 0, 0, 0),
(1, 3, 1000, 1, 98),
(2, 3, 0, 0, 98);

-- ============================================================
-- Cloudinary Image Metadata Table (NEW)
-- ============================================================
CREATE TABLE IF NOT EXISTS `cloudinary_images` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `public_id` VARCHAR(255) NOT NULL,
    `url` VARCHAR(500) NOT NULL,
    `secure_url` VARCHAR(500) NOT NULL,
    `format` VARCHAR(50),
    `width` INT,
    `height` INT,
    `bytes` BIGINT,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `uploaded_by` INT UNSIGNED,
    `folder` VARCHAR(255),
    `tags` VARCHAR(500),
    INDEX idx_public_id (public_id),
    INDEX idx_folder (folder),
    FOREIGN KEY (uploaded_by) REFERENCES `user`(`uid`) ON DELETE SET NULL
);

-- ============================================================
-- Game Images Table (NEW)
-- ============================================================
CREATE TABLE IF NOT EXISTS `game_images` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `game_alias` VARCHAR(255) NOT NULL,
    `image_type` ENUM('thumbnail', 'banner', 'background', 'logo') NOT NULL DEFAULT 'thumbnail',
    `cloudinary_id` INT UNSIGNED NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cloudinary_id) REFERENCES `cloudinary_images`(`id`) ON DELETE CASCADE,
    INDEX idx_game_alias (game_alias),
    INDEX idx_image_type (image_type),
    UNIQUE KEY unique_game_image (game_alias, image_type)
);
