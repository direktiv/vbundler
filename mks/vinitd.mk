.PHONY: vinitd
vinitd: 	## Build vinitd.
vinitd: build-vinitd version-vinitd

.PHONY: version-vinitd
version-vinitd:
	@if [ -d src/vinitd ]; 													\
		then																	\
			printf "vinitd: %s%s\n" `cd src/vinitd && git rev-parse --short HEAD` `cd src/vinitd && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-vinitd
update-vinitd: 		## Re-clone vinitd.
ifneq ("${VINITD}", "skip")
	@rm -f build/vinitd
	@rm -rf src/vinitd
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${VINITD} https://github.com/vorteil/vinitd.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone vinitd branch ${VINITD}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-vinitd
build-vinitd:
ifneq ("${VINITD}", "skip")
	@if [ ! -d src/vinitd ]; 													\
		then																	\
			echo "vinitd has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cd src/vinitd && make all
	@cp src/vinitd/build/vinitd build
	@strip build/vinitd
endif
