build:
	@go build -o out/yt_notifier .

run: build
	./out/yt_notifier
