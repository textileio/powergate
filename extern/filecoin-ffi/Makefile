DEPS:=filecoin.h filecoin.pc libfilecoin.a

all: $(DEPS)
.PHONY: all


$(DEPS): .install-filecoin  ;

.install-filecoin: rust
	./install-filecoin
	@touch $@


clean:
	rm -rf $(DEPS) .install-filecoin
.PHONY: clean
