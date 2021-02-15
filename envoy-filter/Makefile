package:
	docker build -t lua-test:latest .

test: package
	docker run -it -v $(shell pwd):/tests lua-test:latest luajit /tests/filter_test.lua --verbose

dev: 
	docker run -it -v $(shell pwd):/tests lua-test:latest sh

