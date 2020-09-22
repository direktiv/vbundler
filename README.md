<br />
<p align="center">
  <a href="https://github.com/vorteil/vbundler">
    <img src="assets/vlogo.png" alt="Logo" width="80" height="80">
  </a>
  <h3 align="center">vbundler</h3>
  <h5 align="center">build system for vorteil.io micro virtual machines bundles</h5>
</p>

<hr/>

Virtual machines build with [vorteil.io tools](https://github.com/vorteil/vorteil) are based on bundles providing all the required dependencies to run the VM. During build the required files are getting selected from a bundle. This project is the builder for those bundles. This project is used to build the bundle with all dependencies from scratch. If you want to modify [vinitd](https://github.com/vorteil/vinitd) only it is recommended to follow the [vinitd](https://github.com/vorteil/vinitd) documentation of how to modify, run and test the code.

#### Bundle Layout

The bundle is the base for the first partition of the vorteil.io image. It is basically a tar file with an additional metadata file. The following is a shortened example of this metadata file.

**Metadata file**
```yaml
{
  "version": "1.0.0",
  "files": [
  {
    "name": "vinitd",
    "tags": []
  },
  {
    "name": "strace",
    "tags": []
  },
  {
    "name": "fluent-bit",
    "tags": ["logs"]
  },
  {
    "name": "tcpdump",
    "tags": ["tcp"]
    }
  ]
}
```

For the tools to generate the right bundle for an application the items in the bundle are getting a tag. This tag defines if the item gets written to the disk or not. For example if the application uses internal fluentbit logging [vorteil.io tools](https://github.com/vorteil/vorteil) will pick all items tagged as _"logs"_ plus the always required [vinitd](https://github.com/vorteil/vinitd) and linux.

**Bundle to disk partition**
<p align="center">
    <img src="assets/vbundle.png" alt="bundle">
</p>

During build of the image the manifest gets removed and the necessary artifacts picked from the bundle. It is important that linux is the first item in the tar archive. The [bootloader](https://github.com/vorteil/linux-bootloader) loads linux from a fixed offset on the created image.

**Disk Layout**
<p align="center">
    <img src="assets/vdisk.png" alt="bundle">
</p>

The final disk has two partitions. The first one contains the created live bundle and the second partition contains all files from the project used to build this image.

The first partition is mounted under _/vorteil_ during boot of the system. The second partition is mounted under _/_.

**Artifacts for builder:**

- [kernel](https://github.com/vorteil/vlinux)
- [vinitd](https://github.com/vorteil/vinitd)
- [chrony](https://chrony.tuxfamily.org/)
- [fluent-bit](https://github.com/fluent/fluent-bit)
- [fluent-bit disk plugin](https://github.com/vorteil/fluent-bit-disk)
- [busybox](https://busybox.net/)
- [strace](https://github.com/vorteil/strace)
- [tcpdump](https://github.com/vorteil/tcpdump)

#### Building

The build process is supported on Debian and Centos systems. The following commands will create a file in the root directory of the project with the name _kernel-99.99.1_. The version of the bundle can be changed with the BUNDLE_VERSION variable.

```sh
make dependencies
make bundle
```

#### License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.
