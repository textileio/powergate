ASSET_DIRS = $(shell find ./public/ -type d)
ASSET_FILES = $(shell find ./public/ -type f -name '*')

assets.go: ./public/ $(ASSET_DIRS) $(ASSET_FILES)
	go-assets-builder . -p gateway -o assets.go