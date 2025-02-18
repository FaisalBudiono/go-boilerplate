-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';




CREATE TABLE auth_activities (
    id BIGINT PRIMARY KEY NOT NULL GENERATED BY DEFAULT AS IDENTITY,
    user_id BIGINT NOT NULL,
    payload VARCHAR(255) NOT NULL,
    last_activity_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NULL DEFAULT now(),
    updated_at TIMESTAMP NULL DEFAULT now(),
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);




CREATE INDEX idx_auth_activities_user_id ON auth_activities (user_id);




CREATE INDEX idx_auth_activities_payload ON auth_activities (payload);




-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';




DROP TABLE IF EXISTS auth_activities;




-- +goose StatementEnd
