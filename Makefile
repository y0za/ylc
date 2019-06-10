.PHONY: build
build:
	@go build -tags osusergo -ldflags "-X main.googleClientID=$(GOOGLE_CLIENT_ID) -X main.googleClientSecret=$(GOOGLE_CLIENT_SECRET)"
