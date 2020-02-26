package runner

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/projectdiscovery/gologger"
	"github.com/rs/xid"
)

// Runner is a client for running the enumeration process.
type Runner struct {
	tempDir string
	options *Options
}

// New creates a new client for running enumeration process.
func New(options *Options) (*Runner, error) {
	runner := &Runner{
		options: options,
	}

	// Setup the massdns binary path if none was give.
	// If no valid path found, return an error
	if options.MassdnsPath == "" {
		options.MassdnsPath = runner.findBinary()
		if options.MassdnsPath == "" {
			return nil, errors.New("could not find massdns binary")
		}
	}

	// Create a temporary directory that will be removed at the end
	// of enumeration process.
	dir, err := ioutil.TempDir(options.Directory, "shuffledns")
	if err != nil {
		return nil, err
	}
	runner.tempDir = dir

	return runner, nil
}

// Close releases all the resources and cleans up
func (r *Runner) Close() {
	os.RemoveAll(r.tempDir)
}

// findBinary searches for massdns binary in various pre-defined paths
// only linux and macos paths are supported rn
func (r *Runner) findBinary() string {
	locations := []string{
		"/usr/bin/massdns",
		"/usr/local/bin/massdns",
	}

	for _, file := range locations {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			return file
		}
	}
	return ""
}

// runEnumeration sets up the input layer for giving input to massdns
// binary and runs the actual enumeration
func (r *Runner) runEnumeration() error {
	// Handle stdin input
	if r.options.Stdin {
		// Is the stdin input a domain for bruteforce
		if r.options.Wordlist {
			r.processDomain()
		}
		// Write the input from stdin to a file and resolve it.
		r.processSubdomains()
	}

	// Handle a list of subdomains to resolve
	if r.options.SubdomainsList {
		r.processSubdomains()
	}

	// Handle a domain to bruteforce with wordlist
	if r.options.Wordlist {
		r.processDomain()
	}
}

// processDomain processes the bruteforce for a domain using a wordlist
func (r *Runner) processDomain() {
	resolveFile := path.Join(r.tempDir, xid.New().String())
	file, err := os.Create(resolverFile)
	if err != nil {
		gologger.Fatalf("Could not create bruteforce list (%s): %s\n", r.tempDir, err)
		return
	}
	writer := bufio.NewWriter(file)

	// Read the input wordlist for bruteforce generation
	inputFile, err := os.Open(r.options.Wordlist)
	if err != nil {
		gologger.Fatalf("Could not read bruteforce wordlist (%s): %s\n", r.options.Wordlist, err)
		file.Close()
		return
	}

	// Create permutation for domain with wordlist
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		writer.WriteString(text + "." + r.options.Domain)
	}
	writer.Flush()
	inputFile.Close()
	file.Close()

}

// processSubdomain processes the resolving for a list of subdomains
func (r *Runner) processSubdomains() {
	var resolveFile string

	// If there is stdin, write the resolution list to the file
	if r.options.Stdin {
		resolveFile = path.Join(r.tempDir, xid.New().String())
		file, err := os.Create(resolverFile)
		if err != nil {
			gologger.Fatalf("Could not create resolution list (%s): %s\n", tempDir, err)
			return
		}
		io.Copy(file, os.Stdin)
		file.Close()
	} else {
		// Use the file if user has provided one
		resolveFile = r.options.SubdomainsList
	}
}