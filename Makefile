
BUNDLE_VERSION := 99.99.1 	### Kernel bundle version.
COMPILER := 3.0.0 	### Minimum version Vorteil compiler compatible with bundle.
LINKER_DST = "/vorteil/ld-linux-x86-64.so.2"
CC:=gcc
SUDO := sudo

# define version/branches
FLUENTBIT := 'v1.5.4'
BUSYBOX   := '1_32_stable'
CHRONY    := '3.5.1'
STRACE    := 'master'
TCPDUMP   := 'master'
FLUENTDISK:= 'master'
VINITD    := 'master'
# LINUX     := 'v5.8.5'
LINUX     := 'v5.7.10'

BUNDLE_TAGS := --tags=linux \
				--tags=vinitd \
				--tags=strace,strace \
				--tags=tcpdump,tcpdump \
				--tags=fluent-bit,logs \
				--tags=chronyd,ntp \
				--tags=busybox,shell \
				--tags=busybox-install.sh,shell \
				--tags=libnss_dns.so.2,logs,ntp \
				--tags=libnss_files.so.2,logs,ntp \
				--tags=libresolv.so.2,logs,ntp \
				--tags=flb-in_vdisk.so,logs

# enable skipping of components
COPY_FILES_ARGS :=
ifeq ($(FLUENTBIT), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) FLUENTBIT=skip
endif
ifeq ($(BUSYBOX), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) BUSYBOX=skip
endif
ifeq ($(CHRONY), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) CHRONY=skip
endif
ifeq ($(STRACE), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) STRACE=skip
endif
ifeq ($(TCPDUMP), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) TCPDUMP=skip
endif
ifeq ($(FLUENTDISK), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) FLUENTDISK=skip
endif
ifeq ($(VINITD), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) VINITD=skip
endif
ifeq ($(LINUX), skip)
	COPY_FILES_ARGS := $(COPY_FILES_ARGS) LINUX=skip
endif

.PHONY: all
all: 		## Compile all components into a kernel bundle.
all: bundle

.PHONY: bundle
bundle: 	## Compiler all components into a kernel bundle
bundle: bundler busybox vinitd fluent-bit chrony linux strace tcpdump fluentdisk bundle-only

.PHONY: bundle-only
bundle-only: 	## Compiler all components into a kernel bundle
bundle-only:
	@./bundler/build/bundler fetch-libs build
	@./bundler/build/bundler generate-config ${COMPILER} build $(BUNDLE_TAGS) > ./bundler/build/bundle.config && cat ./bundler/build/bundle.config
	@./bundler/build/bundler create ${BUNDLE_VERSION} ./bundler/build/bundle.config > kernel-${BUNDLE_VERSION}

.PHONY: dependencies
dependencies: 	## Clone all dependencies and install required system tools.
dependencies: update
		@if which dnf; then \
			echo "using dnf"; \
			if which go; then echo "Skipping go (already installed)"; else ${SUDO} dnf -y install golang.x86_64; fi; \
			if which g++; then echo "Skipping gcc-c++ (already installed)"; else ${SUDO} dnf -y install gcc-c++; fi; \
			if which gcc; then echo "Skipping gcc (already installed)"; else ${SUDO} dnf -y install gcc; fi; \
			if [ -f /usr/lib64/libcrypt.a ]; then echo "Skipping glibc-static (already installed)"; else ${SUDO} dnf config-manager --enable PowerTools && ${SUDO} dnf install -y glibc-static; fi; \
			if which cmake; then echo "Skipping cmake (already installed)"; else ${SUDO} dnf -y install -y cmake; fi; \
			if which flex; then echo "Skipping flex (already installed)"; else ${SUDO} dnf -y install -y flex; fi; \
			if which bison; then echo "Skipping bison (already installed)"; else ${SUDO} dnf -y install -y bison; fi; \
			if [ ! -f ./libseccomp-2.4.4.tar.gz ]; then \
				${SUDO} dnf install -y libseccomp-devel; \
				wget https://github.com/seccomp/libseccomp/releases/download/v2.4.4/libseccomp-2.4.4.tar.gz; \
				tar -xzf libseccomp-2.4.4.tar.gz; \
				cd libseccomp-2.4.4 && ./configure && make && ${SUDO} make install && cd ..; \
			fi; \
			if [ ! -f ./libpcap-1.9.1.tar.gz ]; then \
				wget http://www.tcpdump.org/release/libpcap-1.9.1.tar.gz; \
				tar -xzf libpcap-1.9.1.tar.gz; \
				cd libpcap-1.9.1 && ./configure && make && ${SUDO} make install; \
			fi; \
			if [ ! -f /usr/include/libelf.h ]; then \
				${SUDO} dnf -y install elfutils-libelf-devel; \
			fi; \
		elif which apt; then \
			echo "using apt"; \
			if which go; then echo "Skipping go (already installed)"; else ${SUDO} apt install -y golang-go; fi; \
			if which gcc; then echo "Skipping gcc (already installed)"; else ${SUDO} apt install -y build-essential; fi; \
			if which cmake; then echo "Skipping cmake (already installed)"; else ${SUDO} apt install -y cmake; fi; \
			if which flex; then echo "Skipping flex (already installed)"; else ${SUDO} apt install -y flex; fi; \
			if which bison; then echo "Skipping bison (already installed)"; else ${SUDO} apt install -y bison; fi; \
			if [ -d /usr/share/doc/libssl-dev ]; then echo "Skipping OpenSSL headers (already installed)"; else ${SUDO} apt install -y libssl-dev; fi; \
			if [ -d /usr/share/doc/libseccomp-dev ]; then echo "Skipping libseccomp headers (already installed)"; else ${SUDO} apt-get install -y libseccomp-dev; fi; \
			if [ -d /usr/share/doc/libpcap-dev ]; then echo "Skipping libpcap-dev headers (already installed)"; else ${SUDO} apt-get install -y libpcap-dev; fi; \
			if [ ! -f /usr/include/libelf.h ]; then \
				${SUDO} apt-get install libelf-dev; \
			fi; \
		else \
			echo "Couldn't find package manager. Skipped installing prerequisite packages."; \
		fi

.PHONY: clean
clean:
		rm -rf build/*
		rm -rf src/*

include mks/*.mk

.PHONY: bundler
bundler: 	## Build fluent-bit.
bundler: build-bundler

.PHONY: build-bundler
build-bundler:
	@cd bundler && mkdir -p build && go build -o build/bundler cmd/main.go

.PHONY: update
update: 	## Re-clone all dependencies.
update: update-fluent-bit update-busybox update-chrony update-strace update-tcpdump update-fluentdisk update-vinitd update-linux

.PHONY: versions
versions:	## List all dependency versions.
versions: version-fluent-bit version-busybox version-chrony version-strace version-tcpdump version-fluentdisk version-vinitd version-linux

.PHONY: dev-vinitd
dev-vinitd: build-vinitd
	@./bundler/build/bundler create ${BUNDLE_VERSION} ./bundler/build/bundle.config > kernel-${BUNDLE_VERSION}
	cp kernel-${BUNDLE_VERSION} ~/.vorteild/kernels/watch/
