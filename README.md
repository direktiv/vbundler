# vbundler

Vorteil virtual machines are based on bundles which contain all binaries and libraries required to run the configured applications. During the build process vorteil extracts the files it needs from those bundles. This project builds these bundles.

Bundles contain:

- [kernel](https://github.com/vorteil/vlinux)
- [vinitd](https://github.com/vorteil/vinitd)
- [chrony](https://chrony.tuxfamily.org/)
- [fluent-bit](https://github.com/fluent/fluent-bit)
- [fluent-bit disk plugin](https://github.com/vorteil/fluent-bit-disk)
- [busybox](https://busybox.net/)
- [strace](https://github.com/vorteil/strace)
- [tcpdump](https://github.com/vorteil/tcpdump)
- shared objects
