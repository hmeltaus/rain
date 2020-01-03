all: lib test

test: rain.c rain.so
	gcc -o $@ $^

lib: rain.h

rain.h: rain.so

rain.so: rain.go
	go build -o $@ -buildmode=c-shared $<

rain.a: rain.go
	go build -o $@ -buildmode=c-archive $<
