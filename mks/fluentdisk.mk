.PHONY: fluentdisk
fluentdisk: 	## Build fluent-bit plugin.
fluentdisk: build-fluentdisk version-fluentdisk

.PHONY: version-fluentdisk
version-fluentdisk:
	@if [ -d src/fluent-bit-disk ]; 													\
		then																	\
			printf "fluent-bit-disk: %s%s\n" `cd src/fluent-bit-disk && git rev-parse --short HEAD` `cd src/fluent-bit-disk && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-fluentdisk
update-fluentdisk:
ifneq ("${FLUENTDISK}", "skip")
	@rm -f build/fluent-bit-disk
	@rm -rf src/fluent-bit-disk
	@mkdir -p src
	@cd src && if git clone --single-branch --branch=${FLUENTDISK} https://github.com/direktiv/fluent-bit-disk.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone fluentdisk branch ${FLUENTDISK}" && \
			exit 1 \
	; \
fi
endif

.PHONY: build-fluentdisk
build-fluentdisk:
ifneq ("${FLUENTDISK}", "skip")
	@if [ ! -d src/fluent-bit-disk ]; 													\
		then																	\
			echo "fluent-bit-disk has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@cd src/fluent-bit-disk && mkdir -p build
	@cd src/fluent-bit-disk/build && cmake -DFLB_SOURCE=../fluent-bit -DPLUGIN_NAME=in_vdisk ../
	@cd src/fluent-bit-disk/build && make
	@cp src/fluent-bit-disk/build/flb-in_vdisk.so build/flb-in_vdisk.so
endif
