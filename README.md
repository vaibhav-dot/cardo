# Cardo

An AI agent for personal finance. It reads your transaction history, finds spending patterns and savings, and suggests relevant deals at the right time.

> Built to learn agent tooling, RAG, evals and prompt engineering with React, TypeScript and Go.

## Layout

```
cardo/
├── ingest/            Go: synthetic data generator (and, from day 3, the analytics API)
│   └── cmd/seed/      writes data/transactions.csv, labelled.csv, answers.json
├── agent/             TypeScript: LLM calls, the agent loop, tools, RAG
│   └── src/hello.ts   day 0: first LLM call with structured output
├── evals/             Promptfoo configs and prompts
├── web/               React + TypeScript UI (create on day 6)
├── db/schema.sql      Postgres + pgvector schema
├── data/              generated data (safe to commit; it's synthetic)
├── docker-compose.yml
└── Makefile
```

## Getting started

Prerequisites: Go 1.22+, Node 20+, Docker Desktop.

```bash
cp .env.example .env    # then add your Anthropic API key

make up        # start Postgres + pgvector
make load      # generate 12 months of fake transactions and load them
make psql      # look around: SELECT * FROM raw_transactions LIMIT 20;

cd agent
npm install @anthropic-ai/sdk dotenv
npm install -D tsx typescript @types/node
cd ..
make hello     # your first structured LLM call

make eval      # categoriser accuracy against the planted answers
```

## Why synthetic data?

`data/answers.json` lists exactly what was planted: 5 subscriptions, a Netflix price rise in March 2026, Disney+ cancelled after May, a Saturday Tesco habit and a £649 one-off spike. Every analytics function and agent answer can be checked against it exactly.

## Eval results

| Component | Metric | v1 | v2 | Notes |
|---|---|---|---|---|
| Categoriser | accuracy | – | – | |
| Agent | correct tool + numbers | – | – | |
| Offer retrieval | recall@5 | – | – | |

## What I'd do next
Open banking sandbox, multi-user auth, agent memory, an MCP server, and backtesting the purchase predictions.
