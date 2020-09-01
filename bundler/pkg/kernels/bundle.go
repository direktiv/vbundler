package kernels

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/vorteil/bundler/pkg/calver"
)

/*
{
	"version": "1.0.0",
	"files": [
		{
			"name": "vkernel-PROD",
			"size": 100000,
			"tags": ["perf", "prod"]
		},
		{
			"name": "vkernel-DEBUG",
			"size": 100000,
			"tags": ["perf", "debug"]
		},
		{
			"name": "bzImage",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "vinitd",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "ld-linux-x86-64.so.2",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "libdl.so.2",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "libfuse.so.2",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "libpthread.so.0",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "libc.so.6",
			"size": 1000000,
			"tags": ["compat"]
		},
		{
			"name": "strace",
			"size": 1000000,
			"tags": ["compat", "debug"]
		},
		{
			"name": "fluent-bit",
			"size": 1000000,
			"tags": ["compat", "logs"]
		}
	]
}

TAGS: perf

TAGS: perf, prod
	vkernel-PROD

TAGS: perf, debug
	vkernel-DEBUG

TAGS: compat
	bzImage
	vinitd
	ld-linux-x86-64.so.2
	libdl.so.2
	libfuse.so.2
	libpthread.so.0
	libc.so.6

TAGS: compat, debug
	bzImage
	vinitd
	ld-linux-x86-64.so.2
	libdl.so.2
	libfuse.so.2
	libpthread.so.0
	libc.so.6
	strace

TAGS: compat, logs
	bzImage
	vinitd
	ld-linux-x86-64.so.2
	libdl.so.2
	libfuse.so.2
	libpthread.so.0
	libc.so.6
	fluent-bit

*/

const ManifestName = "manifest"

type BundleFileMetadata struct {
	Name string   `json:"name"`
	Size int64    `json:"size"`
	Tags []string `json:"tags,omitempty"`
}

type BundleMetadata struct {
	Version                    calver.CalVer        `json:"version"`
	EarliestCompatibleCompiler string               `json:"compiler"`
	Files                      []BundleFileMetadata `json:"files"`
}

func (metadata *BundleMetadata) Marshal() ([]byte, error) {
	return json.Marshal(metadata)
}

func (metadata *BundleMetadata) Unmarshal(data []byte) error {
	return json.Unmarshal(data, metadata)
}

type Bundle struct {
	metadata BundleMetadata
	rs       io.ReadSeeker
}

func NewBundle(rs io.ReadSeeker) (*Bundle, error) {

	var err error
	var gz *gzip.Reader
	var tr *tar.Reader
	var hdr *tar.Header
	var jd *json.Decoder
	var BundleMetadata *Bundle

	_, err = rs.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	gz, err = gzip.NewReader(rs)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr = tar.NewReader(gz)

	hdr, err = tr.Next()
	if err != nil {
		return nil, err
	}

	if hdr.FileInfo().Name() != ManifestName {
		return nil, errors.New("missing manifest data")
	}

	BundleMetadata = new(Bundle)

	jd = json.NewDecoder(tr)

	err = jd.Decode(&BundleMetadata.metadata)
	if err != nil {
		return nil, err
	}

	BundleMetadata.rs = rs

	return BundleMetadata, nil

}

func (bundle *Bundle) Version() calver.CalVer {
	return bundle.metadata.Version
}

func (bundle *Bundle) Files() []BundleFileMetadata {

	var files []BundleFileMetadata
	files = make([]BundleFileMetadata, len(bundle.metadata.Files))

	for i, f := range bundle.metadata.Files {
		files[i] = f
	}

	return files

}

func (bundle *Bundle) EarliestCompatibleCompiler() string {
	return bundle.metadata.EarliestCompatibleCompiler
}

func (bundle *Bundle) Size(tags ...string) int64 {

	var size int64
	var idx int
	var skip bool

	sort.Strings(tags)

	for _, file := range bundle.metadata.Files {
		skip = false
		if len(file.Tags) != 0 {
			hasOrTags := false
			matchesOrTag := false
			for _, tag := range file.Tags {
				isOrTag := false
				if strings.HasPrefix(tag, "+") {
					// different type of tag
					isOrTag = true
					hasOrTags = true
					tag = tag[1:]
				}
				idx = sort.SearchStrings(tags, tag)
				if idx == len(tags) || tags[idx] != tag {
					if !isOrTag {
						skip = true
						break
					}
				} else if hasOrTags {
					matchesOrTag = true
				}
			}
			if hasOrTags && !matchesOrTag {
				skip = true
			}
		}
		if !skip {
			size += (1 + (file.Size+511)/512) * 512
		}
	}

	size += 1024

	return size

}

func (bundle *Bundle) FilesList(tags ...string) []string {

	var files []string
	var skip bool

	sort.Strings(tags)

	for _, file := range bundle.metadata.Files {
		skip = false
		var found bool
		if len(file.Tags) != 0 {
			for _, tag := range file.Tags {
				for _, t := range tags {
					if strings.TrimPrefix(t, "+") == strings.TrimPrefix(tag, "+") {
						found = true
						break
					}
				}
			}
			if !found {
				skip = true
			}
		}
		if !skip {
			files = append(files, file.Name)
		}
	}

	return files
}

func (bundle *Bundle) Reader(tags ...string) io.ReadCloser {

	pr, pw := io.Pipe()

	go func(pw *io.PipeWriter) {

		var err error
		var skip bool
		var hdr *tar.Header

		_, err = bundle.rs.Seek(0, 0)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		gr, err := gzip.NewReader(bundle.rs)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer gr.Close()

		tr := tar.NewReader(gr)

		hdr, err = tr.Next()
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if hdr.Name != ManifestName {
			pw.CloseWithError(errors.New("kernel bundle is corrupt"))
			return
		}

		tw := tar.NewWriter(pw)

		sort.Strings(tags)

		for _, file := range bundle.metadata.Files {

			hdr, err = tr.Next()
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			if hdr.FileInfo().Name() != file.Name || hdr.FileInfo().Size() != file.Size {
				pw.CloseWithError(errors.New("kernel bundle is corrupt"))
				return
			}

			skip = false
			if len(file.Tags) != 0 {
				var found bool
				for _, tag := range file.Tags {
					if found {
						break
					}
					for _, t := range tags {
						if strings.TrimPrefix(t, "+") == strings.TrimPrefix(tag, "+") {
							found = true
							break
						}
					}
				}
				if !found {
					skip = true
				}
			}
			if skip {
				continue
			}

			err = tw.WriteHeader(hdr)
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			_, err = io.Copy(tw, tr)
			if err != nil {
				pw.CloseWithError(err)
				return
			}

		}

		err = tw.Close()
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		pw.Close()

	}(pw)

	return pr

}
