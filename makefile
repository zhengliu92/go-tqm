redis:
	docker run --name redis -p 6379:6379 -d redis:7.2-alpine

.PHONY: redis