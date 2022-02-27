.PHONY: linux
linux: 		## Build linux kernel.
linux: build-linux version-linux

.PHONY: version-linux
version-linux:
	@if [ -d src/linux ]; 													\
		then																	\
			printf "linux: %s%s\n" `cd src/linux && git rev-parse --short HEAD` `cd src/linux && git diff --quiet || echo "-dirty"`	\
		;																		\
	fi

.PHONY: update-linux
update-linux: 		## Re-clone linux kernel.
ifneq ("${LINUX}", "skip")
	@rm -f build/linux
	@rm -rf src/linux
	@rm -rf src/vlinux
	@mkdir -p src
	@cd src && if git clone --depth=5--single-branch --branch=${LINUX} git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git --depth 1; \
	then \
			echo "Successfully cloned repository."  \
		; else \
			echo "Failed to clone linux branch ${LINUX}" && \
			exit 1 \
		; \
	fi
endif

.PHONY: build-linux
build-linux:
ifneq ("${LINUX}", "skip")
	@if [ ! -d src/linux ]; 													\
		then																	\
			echo "linux kernel has not been downloaded (try 'make dependencies' or 'make update')" &&	\
			exit 1																\
		;																		\
	fi
	@if [ ! -d src/linux/drivers/_vorteil ]; 													\
		then																	\
				cd src && rm -Rf vlinux && git clone https://github.com/direktiv/vlinux.git --depth 1 && \
				cd linux && git apply ../vlinux/0001-vorteil-${LINUX}.patch \
		;																		\
	fi
	@if [ ! -d src/linux/.config ]; 													\
		then																	\
				cd src/linux && cp linux.config .config \
		;																		\
	fi
	@cd src/linux && echo "vorteil-$(BUNDLE_VERSION) ($(shell git rev-parse --short HEAD))"
	@cd src/linux && KBUILD_BUILD_USER=vorteil KBUILD_BUILDHOST=vorteil.io KBUILD_BUILD_TIMESTAMP="$(shell date +%d-%m-%Y)" KBUILD_BUILD_VERSION="vorteil-$(BUNDLE_VERSION) ($(shell git rev-parse --short HEAD))" KCFLAGS="-O2 -pipe" make -j8
	@cp src/linux/arch/x86/boot/bzImage build/linux
endif
