# Load .env if present (CRICAPI_KEY for migration 004)
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: deps up down upgrade run daily squad-csv web test test-unit test-component

deps:
	go mod tidy

up:
	docker-compose up -d mysql

down:
	docker-compose down

upgrade: up
	go run ./cmd/upgrade

run:
	go run ./cmd/main

# Daily sync: fetch scorecards for matches missing them, update leaderboard (run via cron)
daily:
	go run ./cmd/daily

# One-off: T20 World Cup squad CSV (name, nationality). Writes to stdout; redirect to file as needed.
squad-csv:
	go run ./cmd/squad-csv

# React app: leaderboard + stats (requires: npm install in web/)
web:
	cd web && npm run dev

test: test-unit test-component

test-unit:
	go test ./bl/ ./endpoints/ ./upgrade/ -v -count=1

test-component:
	MYSQL_DSN="kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true" go test ./dl/mysql/ ./transport/ -v -run "TestComponent|TestHTTP" -count=1
