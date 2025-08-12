CREATE TABLE weighted_tags (
    name VARCHAR(255) NOT NULL,
    weight INT NOT NULL CHECK (weight >= 0)
);