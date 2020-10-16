package cmd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"reflect"
)

var update = flag.Bool("update", false, "Update golden files")
var testBinaryName = "spider-man"

// works out fixture path based on the runtime caller
func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Unable to determine runtime caller data")
	}
	return filepath.Join(filepath.Dir(filename), fixture)
}

// keeps our golden fixtures up to date
func writeFixture(t *testing.T, fixture string, content []byte) {
	err := ioutil.WriteFile(fixturePath(t, fixture), content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// loads a golden fixture from disk
func loadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// start a simple http server to serves some pages for integration tests
func startHttpServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{
		Addr:    ":8000",
		Handler: http.FileServer(http.Dir("./cmd/html/")),
	}

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv
}

// test main function - generates a binary & runs an http server.
func TestMain(m *testing.M) {

	// move up to root project directory
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("Could not change dir: %v", err)
		os.Exit(1)
	}

	// build a test binary for the integration tests
	makeCmd := exec.Command("make", "build-integration")
	err = makeCmd.Run()
	if err != nil {
		fmt.Printf(
			"Could not build test binary for %s: %v",
			testBinaryName,
			err,
		)
		os.Exit(1)
	}

	// start local test server
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	srv := startHttpServer(httpServerExitDone)

	// run the tests
	exitVal := m.Run()

	// shutdown the test server
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	httpServerExitDone.Wait()

	// remove the test binary
	rm := exec.Command("rm", "spider-man")
	err = rm.Run()
	if err != nil {
		fmt.Printf(
			"Could not remove test binary for %s: %v",
			testBinaryName,
			err,
		)
		os.Exit(1)
	}

	// exit based on the exit value
	os.Exit(exitVal)
}

// tests for the CLI commands for crawl
func TestCrawlCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		// base
		{
			"missing base command",
			[]string{},
			"golden/crawl-no-args.golden",
		},
		{
			"with help flag",
			[]string{"crawl", "--help"},
			"golden/crawl-help-flag.golden",
		},
		// depth
		{
			"with depth flag invalid",
			[]string{"crawl", "http://localhost:8000", "--depth", "foo"},
			"golden/crawl-depth-flag-invalid.golden",
		},
		// domain
		{
			"domain and path",
			[]string{"crawl", "http://localhost:8000/foo/bar"},
			"golden/crawl-domain-with-path.golden",
		},
		{
			"domain without scheme",
			[]string{"crawl", "monzo.com"},
			"golden/crawl-domain-without-schema.golden",
		},
		// parallel
		{
			"with parallel flag specified",
			[]string{"crawl", "http://localhost:8000", "--parallel", "1"},
			"golden/crawl-parallel-flag.golden",
		},
		{
			"with parallel flag invalid",
			[]string{"crawl", "http://localhost:8000", "--parallel", "foo"},
			"golden/crawl-parallel-flag-invalid.golden",
		},
		// wait
		{
			"with wait flag missing duration unit",
			[]string{"crawl", "http://localhost:8000", "--wait", "1"},
			"golden/crawl-duration-flag-missing-unit.golden",
		},
		{
			"with wait flag invalid",
			[]string{"crawl", "http://localhost:8000", "--wait", "foo"},
			"golden/crawl-duration-flag-invalid.golden",
		},
		// full run
		// NOTE: We have to pass --parallel 1 here because otherwise we cant guarantee the order
		// of the returned goroutines and so would be unable to compare the returned output with
		// our hardcoded expected output. These are just some simple smoke tests, so not the end of the world.
		{
			"full run with all the things",
			[]string{
				"crawl",
				"http://localhost:8000",
				"--wait", "1s",
				"--parallel", "1",
			},
			"golden/crawl-duration-flag.golden",
		},
	}

	// loop over test cases and run them
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// run the test command against the built binary
			cmd := exec.Command(path.Join(dir, testBinaryName), tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}

			// if update flag is passed, we update the golden fixtures.
			// This is a bit yucky - wonder if there is a cooler way of doing this?
			if *update {
				writeFixture(t, tt.fixture, output)
			}

			// compare actual output with the expected
			actual := string(output)
			expected := loadFixture(t, tt.fixture)

			// if outputs don't match exactly, fail the test
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("actual = %s, expected = %s", actual, expected)
			}
		})
	}
}
