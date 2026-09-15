compose-up:
	docker compose up -d

compose-down:
	docker compose down

image-build:
	docker build -t yvv4docker/browser-chromium .

image-push:
	docker push yvv4docker/browser-chromium:latest

image-pull:
	docker pull yvv4docker/browser-chromium:latest

image-remove:
	docker rmi yvv4docker/browser-chromium:latest

check:
	curl -s http://localhost:9222/json/version
