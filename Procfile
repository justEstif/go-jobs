release: migrate -path ./migrations -database $DATABASE_URL up
web: ./jobs serve
worker: ./jobs scrape --loop
