build:
	docker build -t cage1016/gokitistiok8s-website:latest .

	@if [ "$(PUSH_IMAGE)" == "true" ]; then \
		docker push cage1016/gokitistiok8s-website; \
	fi