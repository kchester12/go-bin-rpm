package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CreatePackage creates an rpm package
func CreatePackage(reposlug, ghToken, email, version, archs, outbuild string, push, keep bool) {

	x := strings.Split(reposlug, "/")
	user := x[0]
	name := x[1]

	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", reposlug)
	fmt.Println("repoPath", repoPath)

	setupGitRepo(repoPath, reposlug, user, email)
	chdir(repoPath)

	if maybesudo(`dnf install rpm-build -y --quiet`) != nil {
		maybesudo(`yum install rpm-build -y --quiet`)
	}

	if tryexec(`latest -v`) != nil {
		exec(`git clone https://github.com/mh-cbon/latest.git %v/src/github.com/mh-cbon/latest`, gopath)
		exec(`go install github.com/mh-cbon/latest`)
	}
	if tryexec(`changelog -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/changelog")
	}
	if tryexec(`go-bin-rpm -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/go-bin-rpm")
	}

	exec(`ls -al %v`, repoPath)
	dir, err := ioutil.TempDir("", "pkg-build")
	if err != nil {
		panic(err)
	}
	for _, arch := range strings.Split(archs, ",") {
		arch = strings.TrimSpace(arch)
		arch = strings.ToLower(arch)
		if arch == "i386" {
			arch = "386"
		} else if arch == "x64" {
			arch = "amd64"
		}

		workDir := filepath.Join(dir, arch)
		outFile := fmt.Sprintf("%v-%v.rpm", name, arch)
		out := filepath.Join(outbuild, outFile)

		mkdirAll(workDir)
		exec(`VERBOSE=* go-bin-rpm generate -a %v --version %v -b %v -o %v`, arch, version, workDir, out)
	}

	exec(`ls -al .`)
	exec(`ls -al %v`, outbuild)
	exec(`ls -al %v`, dir)
	exec(`ls -al %v/*/*`, dir)

	if push == true {
		glob := "*.rpm"
		pushAssetsGh(version, ghToken, outbuild, glob)
		if keep == false {
			exec(`rm -f %v`, outbuild+"/"+glob)
		}
	}

}
