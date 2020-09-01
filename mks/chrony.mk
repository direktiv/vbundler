.PHONY: chrony
chrony: 		## Build linux kernel.
chrony: build-chrony version-chrony

.PHONY: version-chrony
version-chrony:
	@if [ -d src/chrony ]; 													\
		then																	\
			printf "chrony: %s%s\n" `cd src/chrony && git rev-parse --short HEAD` `cd src/chrony && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-chrony
update-chrony: 	## Re-clone fluent-bit.
ifneq ("${CHRONY}", "skip")
	# @rm -f build/fluent-bit
	@rm -rf src/chrony
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${CHRONY} https://github.com/mlichvar/chrony --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone chrony branch ${CHRONY}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-chrony
build-chrony:
ifneq ("${CHRONY}", "skip")
	@if [ ! -d src/chrony ]; 													\
		then																	\
			echo "chrony has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cd src/chrony && make clean || true
	@cd src/chrony &&  LDFLAGS="-Wl,-rpath,/vorteil -Wl,-dynamic-linker,$(LINKER_DST)" ./configure && make
	@mkdir -p build
	@cp src/chrony/chronyd build/chronyd
endif
