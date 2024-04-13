build:
	@go build -o dockerium main.go rootfs.go cgroup.go

run: build
	@./dockerium


.PHONY: build run
