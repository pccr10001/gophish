
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `results` ADD attachment_opened BOOLEAN NOT NULL DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
