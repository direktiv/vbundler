.PHONY: fluent-bit
fluent-bit: 	## Build fluent-bit.
fluent-bit: build-fluent-bit version-fluent-bit

.PHONY: update-fluent-bit
update-fluent-bit: 	## Re-clone fluent-bit.
ifneq ("${FLUENTBIT}", "skip")
	@rm -f build/fluent-bit
	@rm -rf src/fluent-bit
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${FLUENTBIT} https://github.com/fluent/fluent-bit.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
		 	echo "Failed to clone fluentbit branch ${FLUENTBIT}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-fluent-bit
build-fluent-bit:
ifneq ("${FLUENTBIT}", "skip")
	@if [ ! -d src/fluent-bit ]; 													\
		then																	\
			echo "fluent-bit has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@rm -Rf src/fluent-bit/build && mkdir -p src/fluent-bit/build
	@cd src/fluent-bit/build && LDFLAGS="-Wl,-rpath,/vorteil -Wl,-dynamic-linker,$(LINKER_DST)" cmake .. && make
	@mkdir -p build
	@cp src/fluent-bit/build/bin/fluent-bit build/fluent-bit
endif

.PHONY: version-fluent-bit
version-fluent-bit:
	@if [ -d src/fluent-bit ]; 													\
		then																	\
			printf "fluent-bit: %s%s\n" `cd src/fluent-bit && git rev-parse --short HEAD` `cd src/fluent-bit && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi
