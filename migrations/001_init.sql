CREATE TABLE accounts (
  account_id BIGINT PRIMARY KEY,
  balance NUMERIC(20, 5) NOT NULL CHECK (balance >= 0)
);

CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  source_account_id BIGINT NOT NULL,
  destination_account_id BIGINT NOT NULL,
  amount NUMERIC(20, 5) NOT NULL CHECK (amount > 0),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);