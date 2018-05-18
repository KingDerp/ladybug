CREATE TABLE users (
  pk BIGSERIAL PRIMARY KEY,
  fullName VARCHAR NOT NULL
);

CREATE TABLE sessions (
  pk BIGSERIAL PRIMARY KEY,
  user_pk INTEGER REFERENCES users(pk) ON DELETE CASCADE,
  id VARCHAR NOT NULL,
  createDate TIMESTAMP NOT NULL,
  lastUsedDate TIMESTAMP NOT NULL
);