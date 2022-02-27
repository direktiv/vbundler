.PHONY: bundler
bundler: 	## Build bundler.
bundler: build-bundler version-bundler

.PHONY: update-bundler
update-bundler: 	## Re-clone bundler.
ifneq ("${BUNDLER}", "skip")
	@rm -rf src/bundler
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${BUNDLER} https://github.com/direktiv/bundler.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone bundler branch ${BUNDLER}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-bundler
build-bundler:
ifneq ("${BUNDLER}", "skip")
	@if [ ! -d src/bundler ]; 													\
		then																	\
			echo "bundler has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@rm -f bundler/build/bundler
	@mkdir -p bundler/build
	@cd src/bundler && go build -o ../../bundler/build/bundler cmd/main.go
endif

.PHONY: version-bundler
version-bundler:
	@if [ -d src/bundler ]; 													\
		then																	\
			printf "bundler: %s%s\n" `cd src/bundler && git rev-parse --short HEAD` `cd src/bundler && git diff --quiet || echo "-dirty"` \																\
		;																		\
	fi
