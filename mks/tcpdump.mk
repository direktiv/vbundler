.PHONY: tcpdump
tcpdump: 	## Build tcpdump.
tcpdump: build-tcpdump version-tcpdump

.PHONY: version-tcpdump
version-tcpdump:
	@if [ -d src/tcpdump ]; 													\
		then																	\
			printf "tcpdump: %s%s\n" `cd src/tcpdump && git rev-parse --short HEAD` `cd src/tcpdump && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-tcpdump
update-tcpdump: 	## Re-clone tcpdump.
ifneq ("${TCPDUMP}", "skip")
	@rm -f build/tcpdump
	@rm -rf src/tcpdump
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${TCPDUMP} https://github.com/vorteil/tcpdump.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone tcpdump branch ${TCPDUMP}" && \
			exit 1 \
	; \
fi
endif

.PHONY: build-tcpdump
build-tcpdump:
ifneq ("${TCPDUMP}", "skip")
	@if [ ! -d src/tcpdump ]; 													\
		then																	\
			echo "tcpdump has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cd src/tcpdump && CGO_LDFLAGS="-static -w -s -Wl,--dynamic-linker=/vorteil/ld-linux-x86-64.so.2 -Wl,-rpath,/vorteil" go build -v -tags netgo
	@cp src/tcpdump/tcpdump build/tcpdump
endif
