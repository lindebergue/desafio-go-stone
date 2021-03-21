package migrations

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.Register(func(db migrations.DB) error {
		_, err := db.Exec(
			`
				CREATE TABLE IF NOT EXISTS accounts (
					id bigserial PRIMARY KEY,
					name text NOT NULL,
					cpf text NOT NULL,
					secret text NOT NULL,
					balance numeric NOT NULL DEFAULT 0,
					created_at timestamptz NOT NULL DEFAULT now(),
					CHECK (length(name) > 0),
					CHECK (length(cpf) > 0),
					CHECK (length(secret) > 0)
				);

				CREATE UNIQUE INDEX idx_accounts_cpf ON accounts(cpf);

				CREATE TABLE IF NOT EXISTS transfers (
					id bigserial PRIMARY KEY,
					account_origin_id bigint NOT NULL REFERENCES accounts,
					account_destination_id bigint NOT NULL REFERENCES accounts,
					amount numeric NOT NULL,
					created_at timestamptz NOT NULL DEFAULT now()
				);
			`,
		)
		return err
	})
}
