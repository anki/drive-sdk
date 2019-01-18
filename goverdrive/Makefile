GOFILES=$(shell find . -name '*.go')
GOBINDIR=$(shell echo $(GOPATH) |  awk -F: '{ print $$NF }')/bin

.PHONY: build clean save-deps restore-deps test

.DEFAULT_GOAL := all

ALL_EXAMPLES=mover drive sidetap zoneshapes
ALL_GAMES=chicken connect fourmation


ifdef GOBINDIR
clean:
	rm -f $(ALL_EXAMPLES) $(ALL_GAMES)
endif


######################################################################
# TESTS
######################################################################

phystest:
	go test -v -timeout 1m -race github.com/anki/goverdrive/phys/...

tracktest:
	go test -v -timeout 1m -race github.com/anki/goverdrive/robo/track/...

robotest:
	go test -v -timeout 1m -race github.com/anki/goverdrive/robo

test: phystest tracktest robotest


######################################################################
# EXAMPLES
######################################################################

mover: $(GOFILES)
	go build github.com/anki/goverdrive/games/example/mover/

drive: $(GOFILES)
	go build github.com/anki/goverdrive/games/example/drive/

sidetap: $(GOFILES)
	go build github.com/anki/goverdrive/games/example/sidetap/

zoneshapes: $(GOFILES)
	go build github.com/anki/goverdrive/games/example/zoneshapes/

examples: $(ALL_EXAMPLES)


######################################################################
# GAMES
######################################################################

chicken: $(GOFILES)
	go build github.com/anki/goverdrive/games/gwenz/chicken/

connect: $(GOFILES)
	go build github.com/anki/goverdrive/games/gwenz/connect/

fourmation: $(GOFILES)
	go build github.com/anki/goverdrive/games/gwenz/fourmation/

