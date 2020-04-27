CATMGRD_SOURCES_ALL := $(wildcard ./catmgrd/*.go)
CATMGRD_SOURCES := $(filter-out %_test.go, $(CATMGRD_SOURCES_ALL))

build/catmgrd: $(CATMGRD_SOURCES)
	@mkdir -p build
	cd catmgrd; go build
	mv catmgrd/catmgrd build

.PHONY: clean test
clean:
	rm build -rf

test:
	cd catmgrd; go test