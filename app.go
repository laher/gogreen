package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type TestParams struct {
	Pkg     string `json:"pkg"`
	Verbose bool   `json:"verbose"`
	Race    bool   `json:"race"`
}

func (a *App) Chdir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "working directory",
	})
	if err != nil {
		return "", err
	}
	err = os.Chdir(dir)
	return dir, err
}

func (a *App) List() ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}
	lines := []string{"./..."}
	lines = append(lines, strings.Split(string(out), "\n")...)
	lines = lines[:len(lines)-1]
	return lines, err
}

// Run returns a greeting for the given name
func (a *App) Run(p TestParams) (string, error) {
	// count=1 to avoid caching
	params := []string{"test", "-json", "-count=1"}
	if p.Race {
		params = append(params, "-race")
	}
	if p.Verbose {
		params = append(params, "-v")
	}
	params = append(params, p.Pkg)
	cmd := exec.Command("go", params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		runtime.LogWarningf(a.ctx, "err creating stdoutpipe: %s", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		runtime.LogWarningf(a.ctx, "err creating stderrpipe: %s", err)
	}

	err = cmd.Start()
	// stdoutStderr, err := cmd.CombinedOutput()
	// err = nil // ignore for now
	if err != nil {
		return "", err
	}

	f := func(name string, i io.ReadCloser) {
		defer i.Close()
		scanner := bufio.NewScanner(i)
		// optionally, resize scanner's capacity for lines over 64K, see godoc
		for scanner.Scan() {
			runtime.EventsEmit(a.ctx, name, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			runtime.LogWarningf(a.ctx, "err scanning stdoutpipe: %s", err)
		}
	}
	go f("stdout", stdout)
	go f("stderr", stderr)
	//	runtime.LogDebugf(a.ctx, "got %d bytes of data", len(stdoutStderr))
	//	runtime.LogDebugf(a.ctx, "%s", string(stdoutStderr))
	return `{"Action": "test", "Package": "` + p.Pkg + `"}`, nil
}
