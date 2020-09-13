/**
 * SPDX-License-Identifier: Apache-2.0
 * Copyright 2020 vorteil.io Pty Ltd
 */

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/elf"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/vorteil/bundler/pkg/calver"
	"github.com/vorteil/bundler/pkg/kernels"
)

var tagsFlags []string
var processDryRun bool

type FileConfig struct {
	Path string   `toml:"path"`
	Tags []string `toml:"tags"`
}

type Config struct {
	EarliestCompatibleCompiler string       `toml:"compiler"`
	Files                      []FileConfig `toml:"file"`
}

func tarfile(tw *tar.Writer, path, name string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}

	hdr.Name = name

	err = tw.WriteHeader(hdr)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

var cli = &cobra.Command{
	Use:   "bundler",
	Short: "A tool for operating on Vorteil kernel bundles",
}

var create = &cobra.Command{
	Use:   "create VERSION CONFIG",
	Short: "Create a new Vorteil kernel bundle",
	Long: `
Create a new Vorteil kernel bundle with the given VERSION from the provided
CONFIG file.

The CONFIG file is a TOML file describing every file to include in the bundle
and the keywords each should be tagged with:

	compiler = "3.3.0"

	[[file]]
	  path = "/tmp/bzImage"
	  tags = ["compat"]

	[[file]]
	  path = "/tmp/strace"
	  tags = ["compat", "debug"]

The generated kernel bundle is written to stdout, and is not human-readable, so
you probably want to pipe it into a file.`,
	Example: "bundler create 19.09.1 bundle.toml > vorteil-19.09.1",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var config Config
		var bundle kernels.BundleMetadata

		version := args[0]
		path := args[1]

		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = toml.Unmarshal(data, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		gz := gzip.NewWriter(os.Stdout)
		tw := tar.NewWriter(gz)

		bundle.Version, err = calver.Parse(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		for _, fconfig := range config.Files {
			fi, err := os.Stat(fconfig.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			bundle.Files = append(bundle.Files, kernels.BundleFileMetadata{
				Name: filepath.Base(fconfig.Path),
				Size: fi.Size(),
				Tags: fconfig.Tags,
			})
		}
		data, err = bundle.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		f, err := ioutil.TempFile("", "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		_, err = io.Copy(f, bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = tarfile(tw, f.Name(), kernels.ManifestName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		for _, fconfig := range config.Files {
			err = tarfile(tw, fconfig.Path, filepath.Base(fconfig.Path))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}

		err = tw.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = gz.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	},
}

var extract = &cobra.Command{
	Use:   "extract BUNDLE DEST",
	Short: "Extract the contents of a kernel bundle into a directory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bundle := args[0]
		dest := args[1]

		f, err := os.Open(bundle)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		gr, err := gzip.NewReader(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer gr.Close()

		tr := tar.NewReader(gr)

		hdr, err := tr.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if hdr.Name != kernels.ManifestName {
			fmt.Fprintf(os.Stderr, "error: not a valid kernel bundle\n")
			os.Exit(1)
		}

		data, err := ioutil.ReadAll(tr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		metadata := new(kernels.BundleMetadata)
		err = metadata.Unmarshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = os.MkdirAll(dest, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		config := new(Config)

		for i := 0; true; i++ {
			hdr, err = tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			if hdr.Name != metadata.Files[i].Name {
				fmt.Fprintf(os.Stderr, "error: kernel bundle manifest is corrupt\n")
				os.Exit(1)
			}

			path := filepath.Join(dest, hdr.Name)

			config.Files = append(config.Files, FileConfig{
				Path: path,
				Tags: metadata.Files[i].Tags,
			})

			f, err = os.Create(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			_, err = io.Copy(f, tr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			err = f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}

		buf := new(bytes.Buffer)
		enc := toml.NewEncoder(buf)
		err = enc.Encode(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		path := filepath.Join(dest, "bundle.toml")
		_, err = os.Stat(path)
		if !os.IsNotExist(err) {
			if err == nil {
				fmt.Fprintf(os.Stderr, "error: skipping bundle.toml because it already exists\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile(path, buf.Bytes(), 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	},
}

var inspect = &cobra.Command{
	Use:   "inspect BUNDLE",
	Short: "Scan a kernel bundle and print its metadata to stdout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		bundle, err := kernels.NewBundle(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Version: %s\n", bundle.Version())
		fmt.Printf("Earliest Compatible Compiler: %s\n", bundle.EarliestCompatibleCompiler())
		fmt.Printf("Files:\n")
		for _, x := range bundle.Files() {
			fmt.Printf("  %s\n    %v\n", x.Name, x.Tags)
		}
	},
}

var process = &cobra.Command{
	Use:   "process BUNDLE TAG...",
	Short: "Apply tags to a kernel bundle to generate the runtime kernel tar",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		tags := args[1:]

		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		bundle, err := kernels.NewBundle(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if processDryRun {
			list := bundle.FilesList(tags...)
			for _, s := range list {
				fmt.Println(s)
			}
			return
		}

		rc := bundle.Reader(tags...)
		defer rc.Close()

		_, err = io.Copy(os.Stdout, rc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
}

var fetchLibs = &cobra.Command{
	Use:    "fetch-libs DIR",
	Short:  "Scan DIR for 64-bit ELF files and fetch all needed shared objects.",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		opdir := args[0]

		pthChan := make(chan string)
		errChan := make(chan error)
		finChan := make(chan error)

		fetched := make(map[string]bool)
		required := make(map[string]bool)

		go func() {
			var getFilepaths func(dir string)
			getFilepaths = func(dir string) {
				fs, err := ioutil.ReadDir(dir)
				if err != nil {
					errChan <- err
					return
				}

				for _, f := range fs {
					p := filepath.Join(dir, f.Name())
					if f.IsDir() {
						getFilepaths(p)
						continue
					}

					pthChan <- p
				}
			}

			getFilepaths(opdir)
			errChan <- nil
		}()

		go func() {
			var errNoContinue = fmt.Errorf("do not continue")
			var mapLock sync.Mutex
			var libFolders = []string{"/lib", "/lib64", "/usr/lib"}

			getImportedLibraries := func(path string) ([]string, error) {
				f, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				defer f.Close()

				e, err := elf.NewFile(f)
				if err != nil {
					// not an elf file, continue
					return nil, nil
				}
				defer e.Close()
				return e.ImportedLibraries()
			}

			fetchSharedObject := func(p string) error {
				if _, ok := fetched[p]; ok {
					return nil
				}
				required[p] = true

				for _, libFolder := range libFolders {
					fi, err := os.Lstat(libFolder)
					if err != nil {
						continue
					}

					if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
						libFolder, err = os.Readlink(libFolder)
						if err != nil {
							continue
						}
						if !strings.HasPrefix(libFolder, "/") {
							libFolder = "/" + libFolder
						}
					}

					err = filepath.Walk(libFolder, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						if strings.Contains(path, "vmware") {
							return nil
						}

						if filepath.Base(path) == p {

							e, err := elf.Open(path)
							if err != nil {
								return err
							}
							defer e.Close()

							if e.Class != elf.ELFCLASS64 {
								return nil
							}

							err = func() error {
								mapLock.Lock()
								defer mapLock.Unlock()
								fetched[p] = true

								var lib = "lib"
								if p == "ld-linux-x86-64.so.2" {
									lib = "lib64"
								}
								lib = "."

								localPath := filepath.Join(opdir, lib, p)
								fmt.Fprintf(os.Stderr, "Copying: %s -> %s\n", p, localPath)
								err := os.MkdirAll(filepath.Dir(localPath), 0777)
								if err != nil {
									return err
								}

								f, err := os.Create(localPath)
								if err != nil {
									return err
								}
								defer f.Close()

								x, err := os.Open(path)
								if err != nil {
									return err
								}
								defer x.Close()

								_, err = io.Copy(f, x)
								if err != nil {
									return err
								}

								return nil
							}()
							if err != nil {
								return err
							}

							go func() {
								pthChan <- path
							}()

							return errNoContinue
						}

						return nil
					})
					if err != nil && err != errNoContinue {
						return err
					}
				}

				return nil
			}

			// always needed for name resolution
			fetchSharedObject("libnss_dns.so.2")
			fetchSharedObject("libnss_files.so.2")
			fetchSharedObject("libresolv.so.2")

			for {
				select {
				case p := <-pthChan:

					err := func() error {
						libs, err := getImportedLibraries(p)
						if err != nil {
							return err
						}

						for _, l := range libs {
							err = fetchSharedObject(l)
							if err != nil {
								return err
							}
						}

						return nil
					}()
					if err != nil {
						go func() {
							errChan <- err
						}()
					}

				case x := <-errChan:
					finChan <- x
					return
				}
			}
		}()

		err := <-finChan
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
			return
		}

		// 				--tags=libnss_dns.so.2,logs,ntp \
		// --tags=libnss_files.so.2,logs,ntp \
		// --tags=libresolv.so.2,logs,ntp \

		var warnings int
		for p, _ := range required {
			if _, ok := fetched[p]; !ok {
				warnings++
				fmt.Fprintf(os.Stderr, "WARNING: Unable to locate shared object: '%s'\n", p)
			}
		}
	},
}

var generateConfig = &cobra.Command{
	Use:    "generate-config COMPILER DIR",
	Short:  "Scan DIR for files and generate a config file for them.",
	Args:   cobra.ExactArgs(2),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		compiler := args[0]
		path := args[1]

		config := new(Config)
		config.EarliestCompatibleCompiler = compiler

		fis, err := ioutil.ReadDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		var binaries []string
		var libs []string
		var misc []string

		m := make(map[string][]string)
		for _, fi := range fis {
			m[fi.Name()] = []string{}

			ef, err := elf.Open(filepath.Join(path, fi.Name()))
			if err != nil {
				misc = append(misc, fi.Name())
				continue
			}

			switch ef.FileHeader.Type {
			case elf.ET_EXEC:
				binaries = append(binaries, fi.Name())
			case elf.ET_DYN:
				libs = append(libs, fi.Name())
			default:
				fmt.Fprintf(os.Stderr, "warning: unknown ELF header type %v for file %s\r\tthis file will be included on every disk\n", ef.FileHeader.Type, fi.Name())
				misc = append(misc, fi.Name())
			}
			ef.Close()
		}

		for _, flag := range tagsFlags {
			strs := strings.Split(flag, ",")
			k := strs[0]
			v, ok := m[k]
			if !ok {
				err = fmt.Errorf("no such file found in DIR: %s", k)
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			v = append(v, strs[1:]...)
			m[k] = v
		}

		// scan path for files and determine which ones are shared objects
		var traversed map[string]bool
		var recurser func(string, ...string)
		recurser = func(name string, tags ...string) {

			if _, ok := traversed[name]; ok {
				return
			}

			ef, err := elf.Open(filepath.Join(path, name))
			if err != nil {
				panic(err)
			}
			defer ef.Close()

			traversed[name] = true

			v := m[name]
			if len(v) == 0 {
				v = tags
			} else {
				var squashed bool

				for _, s := range v {
					if strings.HasPrefix(s, "+") {
						squashed = true
						break
					}
				}

				sort.Strings(v)
				sort.Strings(tags)
				var b bool
				var intersectable = true
				var common = make(map[string]bool)
				var left string
				var right string

				// subset
				b = false
				for _, s := range v {
					idx := sort.SearchStrings(tags, s)
					if idx >= len(tags) || tags[idx] != s {
						// not a subset
						b = true
						if left == "" {
							left = "+" + s
						} else {
							intersectable = false
							break
						}
					} else {
						common[s] = true
					}
				}
				if !b {
					goto set
				}

				// superset
				b = false
				for _, s := range tags {
					idx := sort.SearchStrings(v, s)
					if idx >= len(v) || v[idx] != s {
						// not a superset
						b = true
						if right == "" {
							right = "+" + s
						} else {
							intersectable = false
							break
						}
					} else {
						common[s] = true
					}
				}
				if !b {
					v = tags
					goto set
				}

				if squashed {
					// ERROR SUPPRESSION
					//fmt.Fprintf(os.Stderr, "warning: tag merging failure for file %s\n\tthis file will be included on every disk\n", name)
					//fmt.Fprintf(os.Stderr, "\tv %v tags %v\n", v, tags)
					v = []string{}
					goto set
				}

				// intersection
				if !intersectable {
					// ERROR SUPPRESSION
					//fmt.Fprintf(os.Stderr, "warning: tag merging intersection failure for file %s\n\tthis file will be included on every disk\n", name)
					//fmt.Fprintf(os.Stderr, "v: %v; tags: %v\n", v, tags)
					v = []string{}
					goto set
				}

				v = []string{}
				for k := range common {
					v = append(v, k)
				}
				v = append(v, left, right)
				goto set
			}

		set:
			m[name] = v

			// scan program headers and recurse into them
			sos, err := ef.ImportedLibraries()
			if err != nil {
				panic(err)
			}
			for _, so := range sos {
				so = filepath.Base(so)
				if _, ok := m[so]; !ok {
					fmt.Fprintf(os.Stderr, "warning: detected a possible missing shared object file %s\n", so)
					continue
				}
				recurser(so, tags...)
			}
		}

		for _, bin := range binaries {
			tags := m[bin]
			if len(tags) == 0 {
				fmt.Fprintf(os.Stderr, "warning: binary file has no defined tags %s\n\tthis file will be included on every disk\n", bin)
				continue
			}
			traversed = make(map[string]bool)
			recurser(bin, tags...)
		}

		for _, so := range libs {
			tags := m[so]
			if len(tags) == 0 {
				continue
			}
			traversed = make(map[string]bool)
			recurser(so, tags...)
		}

		// add specified files to config in the order they were specified
		for _, flag := range tagsFlags {
			strs := strings.Split(flag, ",")
			k := strs[0]
			v := m[k]
			config.Files = append(config.Files, FileConfig{
				Path: filepath.Join(path, k),
				Tags: v,
			})
			delete(m, k)
		}

		// alphabetize these remaining files to ensure consistent builds
		var strs []string
		for k := range m {
			strs = append(strs, k)
		}
		sort.Strings(strs)

		for _, k := range strs {
			v := m[k]
			config.Files = append(config.Files, FileConfig{
				Path: filepath.Join(path, k),
				Tags: v,
			})
		}

		buf := new(bytes.Buffer)
		enc := toml.NewEncoder(buf)
		err = enc.Encode(*config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		data := buf.Bytes()

		// "+" SUPPRESSION
		io.Copy(os.Stdout, strings.NewReader(strings.ReplaceAll(string(data), "+", "")))
	},
}

func main() {
	fs := process.Flags()
	fs.BoolVar(&processDryRun, "dry-run", false, "List files that would be included instead of building an archive.")

	fs = generateConfig.Flags()
	fs.StringArrayVar(&tagsFlags, "tags", nil, "Apply requirement tags to a file with format --tags=FILE,TAG[,TAG...]")

	cli.AddCommand(create)
	cli.AddCommand(extract)
	cli.AddCommand(inspect)
	cli.AddCommand(process)
	cli.AddCommand(fetchLibs)
	cli.AddCommand(generateConfig)

	err := cli.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, cli.UsageString())
		os.Exit(1)
	}
}
