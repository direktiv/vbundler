.PHONY: strace
strace: 	## Build strace.
strace: build-strace version-strace

.PHONY: version-strace
version-strace:
	@if [ -d src/strace ]; 													\
		then																	\
			printf "strace: %s%s\n" `cd src/strace && git rev-parse --short HEAD` `cd src/strace && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-strace
update-strace: 	## Re-clone strace.
ifneq ("${STRACE}", "skip")
	@rm -f build/strace
	@rm -rf src/strace
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${STRACE} git@github.com:vorteil/strace.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone strace branch ${STRACE}" && \
			exit 1 \
	; \
fi
endif

.PHONY: build-strace
build-strace:
ifneq ("${STRACE}", "skip")
	@if [ ! -d src/strace ]; 													\
		then																	\
			echo "strace has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cd src/strace && CGO_LDFLAGS="-static -w -s -Wl,--dynamic-linker=/vorteil/ld-linux-x86-64.so.2 -Wl,-rpath,/vorteil" go build -tags netgo
	@cp src/strace/strace build/strace
endif
