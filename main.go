package main

import (
	"flag"
	"fmt"
	"log"

	"semtag/pkg/docker"
	"semtag/pkg/git"
	"semtag/pkg/version"
)

var (
	dryRun bool
	in     string
	out    string
	suffix string
	prefix string
)
var (
	tagGit    bool
	tagDocker bool
	tagFile   bool
)

func parseFlags() {
	flag.BoolVar(&dryRun, "dry-run", false, "if true, only print the object(s) that would be sent, without sending the data")

	flag.StringVar(&in, "in", "", `input data: can be either 1) a docker image .tar file without the file extension (e.g. "api") or 2) a file that contains the version number (e.g. "setup.py")`)
	flag.StringVar(&out, "out", "", `output: can be either 1) a docker repository or 2) the pattern for the file version (e.g. "version='%s',")`)
	flag.StringVar(&suffix, "suffix", "", `if set, append the suffix to the version number (e.g. "0.1.0-rc")`)
	flag.StringVar(&prefix, "prefix", "", `if set, append the prefix to the version number (e.g. "api-0.1.0")`)

	flag.BoolVar(&tagGit, "git", false, "tag git commit")
	flag.BoolVar(&tagDocker, "docker", false, "tag docker image")
	flag.BoolVar(&tagFile, "file", false, "update the version number in a file")

	flag.Parse()

	if dryRun {
		log.Println("dry run mode enabled")
	}
}

func main() {
	parseFlags()

	var ver, nextVer version.Version
	ver.Suffix = suffix
	ver.Prefix = prefix
	ver = *ver.GetLatest()
	nextVer = *ver.GetLatest()
	changeType := nextVer.IncrementAuto()
	log.Println("current version:", ver.String())
	log.Println("next version:", nextVer.String())

	if tagFile {
		f := version.File{
			Path:          in,
			VersionFormat: out,
			Version:       nextVer.String(),
		}
		newContents := f.ReplaceSubstring()
		if !dryRun {
			f.Write(newContents)
			git.Add(in)
			git.Commit(fmt.Sprintf("set version %s %s in %s", nextVer.String(), changeType.String(), in))
			git.Push("")
		}
	}

	if tagGit {
		tag := &git.TagObj{
			Name: nextVer.String(),
		}
		tag.SetMessage()
		log.Println(tag)
		if !dryRun {
			tag.Push()
		}
	}

	if tagDocker {
		docker.Load(in + ".tar")
		img := &docker.Image{
			Name:                in,
			Tags:                ver.AsList(),
			ContainerRepository: out,
		}
		log.Println(img)
		if !dryRun {
			img.Tag()
			img.Push()
		}
	}

}
