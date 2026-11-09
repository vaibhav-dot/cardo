-include .env
export

.PHONY: up down seed load psql hello eval reset

up:            ## start Postgres + pgvector
	docker compose up -d

down:
	docker compose down

seed:          ## generate synthetic data into ./data
	cd ingest && go run ./cmd/seed -out ../data

load: seed     ## load transactions.csv into Postgres
	docker compose exec -T db psql -U cardo -d cardo -c "TRUNCATE raw_transactions;"
	docker compose exec -T db psql -U cardo -d cardo -c "\copy raw_transactions(date, description, amount) FROM '/data/transactions.csv' CSV HEADER"
	docker compose exec -T db psql -U cardo -d cardo -c "SELECT count(*) AS rows_loaded FROM raw_transactions;"

psql:          ## open a SQL shell
	docker compose exec db psql -U cardo -d cardo

hello:         ## run your first LLM call (structured output)
	cd agent && npm run hello

eval:          ## run the categoriser eval
	cd evals && npx promptfoo@latest eval

reset:         ## wipe the database volume
	docker compose down -v
