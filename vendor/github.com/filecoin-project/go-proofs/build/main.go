package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var lineBreak = "\n"

func init() {
	log.SetFlags(0)
	if runtime.GOOS == "windows" {
		lineBreak = "\r\n"
	}
}

// run executes a given command on the shell, like
// `run("git status")`
func run(name string) string {
	args := strings.Split(name, " ")
	return runParts(args...)
}

func runParts(args ...string) string {
	name := strings.Join(args, " ")
	cmd := exec.Command(args[0], args[1:]...) // #nosec
	log.Println(name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", out)
		log.Fatalf("Command '%s' failed: %s\n", name, err)
	}

	return strings.Trim(string(out), lineBreak)
}

// deps installs all dependencies
func deps() {
	log.Println("Installing dependencies...")

	list := []string{
		"go get -u github.com/whyrusleeping/gx",
		"go get -u github.com/whyrusleeping/gx-go",
		"gx install",
		"go get -u github.com/alecthomas/gometalinter",
		"gometalinter --install",
		"go get -u github.com/stretchr/testify",
	}

	for _, name := range list {
		log.Println(run(name))
	}
}

// lint runs linting using gometalinter
func lint(packages ...string) {
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	log.Printf("Linting %s ...\n", strings.Join(packages, " "))

	// Run fast linters batched together
	configs := []string{
		"gometalinter",
		"--skip=sharness",
		"--skip=vendor",
		"--disable-all",
	}

	fastLinters := []string{
		"--enable=vet",
		"--enable=gofmt",
		"--enable=misspell",
		"--enable=goconst",
		"--enable=golint",
		"--enable=errcheck",
		"--min-occurrences=6", // for goconst
	}

	log.Println(runParts(append(append(configs, fastLinters...), packages...)...))

	// FIXME: re-enable commented linters at some point.
	slowLinters := []string{
		"--deadline=10m",
		"--enable=unconvert",
		"--enable=gosimple",
		//"--enable=megacheck",
		"--enable=varcheck",
		//"--enable=structcheck",
		"--enable=deadcode",
	}

	log.Println(runParts(append(append(configs, slowLinters...), packages...)...))
}

var libraryPackages = []string{
	"proofs",
	"porep",
	"porep/drgporep",
	"post/batchedpost",
	"snark",
	"snarkwrapper",
}

func build() {
	log.Println("Building ...")

	for _, p := range libraryPackages {
		buildLibraryPackage(p)
	}
}

func buildLibraryPackage(path string) {
	p := "./" + path //filepath.Join("'", path)

	log.Printf("Building %s ...\n", p)

	commit := run("git log -n 1 --format=%H")

	log.Println(
		runParts(
			"go", "build",
			"-ldflags", fmt.Sprintf("-X github.com/filecoin-project/go-proofs/flags.Commit=%s", commit),
			"-v", p,
		),
	)
}

//func buildProgram(path string, name string) {
//	p := "./" + path
//	log.Printf("Building %s as %s ...\n", p, name)
//
//	commit := run("git log -n 1 --format=%H")
//
//	nm := filepath.Join(path, name)
//
//	log.Println(
//		runParts(
//			"go", "build",
//			"-ldflags", fmt.Sprintf("-X github.com/filecoin-project/go-proofs/flags.Commit=%s", commit),
//			"-v", "-o", nm, p,
//		),
//	)
//}

func install() {
	log.Println("Installing...")

	log.Println(runParts("go", "install"))
}

// test executes tests and passes along all additional arguments to `go test`.
func test(args ...string) {
	log.Println("Testing...")

	log.Println(run(fmt.Sprintf("go test ./... %s", strings.Join(args, " "))))
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalf("Missing command")
	}

	cmd := args[0]

	switch cmd {
	case "deps":
		deps()
	case "lint":
		lint(args[1:]...)
	case "build":
		build()
	case "test":
		test(args[1:]...)
	case "install":
		install()
	case "best":
		build()
		test(args[1:]...)
	case "all":
		deps()
		lint()
		build()
		test(args[1:]...)
	default:
		log.Fatalf("Unknown command: %s\n", cmd)
	}
}