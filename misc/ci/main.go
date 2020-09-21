package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Manifest struct {
	VCLI       Version   `yaml:"vcli"`
	Bootloader Version   `yaml:"bootloader"`
	Trampoline Version   `yaml:"trampoline"`
	Kernels    []Version `yaml:"kernels"`
}
type Version struct {
	Version string `yaml:"version"`
	Release string `yaml:"release"`
}

var (
	tagFlag          string
	kernelSourceFlag string
)

func init() {
	flag.StringVar(&tagFlag, "tag", "", "tag used to version kernels")
	flag.StringVar(&kernelSourceFlag, "kernel-source", "", "kernel-source to upload kernels to and download manifest from")
	flag.Parse()
	if tagFlag == "" || kernelSourceFlag == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	t := time.Now().UTC()
	time := fmt.Sprintf("%s.%vZ", t.Format("2006-01-02T15:04:05"), t.Nanosecond())

	resp, err := http.Get(fmt.Sprintf("https://downloads.vorteil.io/%s/manifest.txt", kernelSourceFlag))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	manifest := new(Manifest)
	err = yaml.Unmarshal(b, &manifest)
	if err != nil {
		log.Fatal(err)
	}

	manifest.Kernels = append(manifest.Kernels, Version{
		Version: tagFlag,
		Release: time,
	})

	f, err := os.Create("manifest.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err = yaml.Marshal(manifest)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(f, bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}

}
