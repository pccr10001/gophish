
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `results` ADD COLUMN attachment_opened BOOLEAN;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
