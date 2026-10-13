compose-up:
	docker compose up -d

compose-down:
	docker compose down

image-build:
	docker build -f docker/default/Dockerfile -t yvv4docker/browser-chromium:default .

image-build-optimized:
	docker build -f docker/optimized/Dockerfile -t yvv4docker/browser-chromium:optimized .

image-build-selkies:
	docker build -f docker/selkies/Dockerfile -t yvv4docker/browser-chromium:selkies .

image-push:
	docker push yvv4docker/browser-chromium:latest

image-pull:
	docker pull yvv4docker/browser-chromium:latest

image-remove:
	docker rmi yvv4docker/browser-chromium:default yvv4docker/browser-chromium:optimized yvv4docker/browser-chromium:selkies

check:
	curl -s http://localhost:9222/json/version
