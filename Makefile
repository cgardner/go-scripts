build: build-tasks

build-tasks: clean-tasks
	go build -mod=vendor -o dist/tasks cmd/tasks/tasks.go

clean-tasks:
	rm -rf dist/tasks

install: build
	cp dist/* /usr/local/bin
