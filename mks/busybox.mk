.PHONY: busybox
busybox: 	## Build busybox.
busybox: build-busybox version-busybox

.PHONY: update-busybox
update-busybox: 	## Re-clone busybox.
ifneq ("${BUSYBOX}", "skip")
	@rm -f build/busybox
	@rm -rf src/busybox
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${BUSYBOX} https://github.com/mirror/busybox --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone busybox branch ${BUSYBOX}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-busybox
build-busybox:
ifneq ("${BUSYBOX}", "skip")
	@if [ ! -d src/busybox ]; 													\
		then																	\
			echo "busybox has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cp misc/busybox.config src/busybox/.config
	@cd src/busybox && make -j4
	@mkdir -p build
	@cp src/busybox/busybox build/busybox
endif

.PHONY: version-busybox
version-busybox:
	@if [ -d src/busybox ]; 													\
		then																	\
			printf "busybox: %s%s\n" `cd src/busybox && git rev-parse --short HEAD` `cd src/busybox && git diff --quiet || echo "-dirty"` \																\
		;																		\
	fi
