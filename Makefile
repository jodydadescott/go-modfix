# cSpell:ignore modfix GOPATH trimpath

build:
	env CGO_ENABLED=0 go build -v -trimpath
	cp modfix $(GOPATH)/bin

clean:
	$(RM) modfix