build: build-tasks build-finance

build-tasks: clean-tasks
	go build -mod=vendor -o dist/tasks cmd/tasks/tasks.go

build-finance: clean-finance
	go build -mod=vendor -o dist/finance cmd/finance/finance.go

clean-tasks:
	rm -rf dist/tasks

clean-finance:
	rm -rf dist/finance

install: build
	cp dist/* /usr/local/bin
