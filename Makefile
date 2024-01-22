GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

tools:
	rm -rf bin/*
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-mdparse cmd/wof-mdparse/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-md2feed cmd/wof-md2feed/main.go	
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-md2html cmd/wof-md2html/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-md2idx cmd/wof-md2idx/main.go

dist-build:
	OS=darwin make dist-os
	OS=windows make dist-os
	OS=linux make dist-os

dist-os:
	mkdir -p dist/$(OS)
	GOOS=$(OS) GOARCH=386 go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o dist/$(OS)/wof-mdparse cmd/wof-mdparse/main.go
	GOOS=$(OS) GOARCH=386 go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o dist/$(OS)/wof-md2feed cmd/wof-md2feed/main.go
	GOOS=$(OS) GOARCH=386 go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o dist/$(OS)/wof-md2html cmd/wof-md2html/main.go
	GOOS=$(OS) GOARCH=386 go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o dist/$(OS)/wof-md2idx cmd/wof-md2idx/main.go
