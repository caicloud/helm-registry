GO = go
GOBUILD = $(GO) build
DEST = ./bin
TEST = ./e2etest
TEST_BUILD = $(TEST)/registry

.PHONY : registry
registry : 
	$(GOBUILD) -o $(DEST)/$@ ./cmd/registry

.PHONY : testbuild
testbuild:
	$(GOBUILD) -race -o $(TEST_BUILD) ./cmd/registry

.PHONY : e2etest
e2etest :
	bash ./test/e2e.sh $(TEST_BUILD)

.PHONY : test
test : | testbuild e2etest clean
	
.PHONY : clean
clean :
	rm -rf $(TEST)

