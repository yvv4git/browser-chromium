compose-up:
	docker compose up -d

compose-down:
	docker compose down

image-build-default:
	docker build -f docker/default/Dockerfile -t yvv4docker/browser-chromium:default .

image-build-optimized:
	docker build -f docker/optimized/Dockerfile -t yvv4docker/browser-chromium:optimized .

image-build-selkies:
	docker build -f docker/selkies/Dockerfile -t yvv4docker/browser-chromium:selkies .

image-build-headless:
	docker build -f docker/headless/Dockerfile -t yvv4docker/browser-chromium:headless .

image-push:
	docker push yvv4docker/browser-chromium:latest

image-pull:
	docker pull yvv4docker/browser-chromium:latest

image-remove:
	docker rmi yvv4docker/browser-chromium:default yvv4docker/browser-chromium:optimized yvv4docker/browser-chromium:selkies yvv4docker/browser-chromium:headless

check:
	curl -s http://localhost:9222/json/version
