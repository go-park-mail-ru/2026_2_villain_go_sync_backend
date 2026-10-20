ALTER TABLE vacancy_history
	ADD COLUMN currency TEXT NOT NULL DEFAULT 'RUB',
	ADD CONSTRAINT vacancy_currency_allowed
		CHECK (currency IN ('RUB', 'USD', 'EUR'));
