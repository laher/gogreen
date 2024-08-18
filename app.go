package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	watching bool
	watcher  *fsnotify.Watcher
	stop     chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{stop: make(chan struct{})}
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

type State struct {
	CWD      string   `json:"cwd"`
	PkgList  []string `json:"pkg_list"`
	Watching bool     `json:"watching"`
}

func (a *App) List() (State, error) {
	cmd := exec.Command("go", "list", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return State{}, err
	}
	lines := []string{"./..."}
	lines = append(lines, strings.Split(string(out), "\n")...)
	lines = lines[:len(lines)-1]
	wd, err := os.Getwd()
	if err != nil {
		return State{}, err
	}
	return State{CWD: wd, PkgList: lines, Watching: a.watching}, err
}

func (a *App) fsnotify(p TestParams) (*fsnotify.Watcher, error) {
	if a.watching || a.watcher != nil {
		return nil, errors.New("already watching")
	}
	var err error
	a.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	a.watcher.Add(".")
	a.watching = true

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				runtime.LogDebugf(a.ctx, "watch event: %+v", event)
				// TODO - debounce?
				s, err := a.Run(p)
				if err != nil {

					runtime.LogErrorf(a.ctx, "error running tests: %s", err)
					continue
				}
				runtime.LogDebugf(a.ctx, "ran tests: %s", s)
			case err, ok := <-a.watcher.Errors:
				if !ok {
					return
				}
				runtime.LogErrorf(a.ctx, "watch error: %s", err)
			case <-a.stop:
				runtime.LogDebugf(a.ctx, "stop watching")
				a.watcher = nil
				a.watching = false
				return
			}
		}
	}()
	return a.watcher, nil
}

func (a *App) Watch(p TestParams) (string, error) {
	_, err := a.fsnotify(p)
	if err != nil {
		return "", err
	}
	// run once to start things off
	return a.Run(p)
}

func (a *App) Unwatch() error {
	if !a.watching {
		return errors.New("not watching")
	}
	a.stop <- struct{}{}
	return nil
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
	runtime.EventsEmit(a.ctx, "cls")

	f := func(name string, i io.ReadCloser) {
		defer i.Close()
		scanner := bufio.NewScanner(i)
		// optionally, resize scanner's capacity for lines over 64K, see godoc
		for scanner.Scan() {
			runtime.EventsEmit(a.ctx, name, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			runtime.LogWarningf(a.ctx, "err scanning %s: %s", name, err)
			runtime.EventsEmit(a.ctx, "err:scan."+name, err)
		}
		runtime.EventsEmit(a.ctx, "done."+name)
	}
	go f("stdout", stdout)
	go f("stderr", stderr)
	go func() {
		err := cmd.Wait()
		if err != nil {
			// TODO publish some error
			runtime.EventsEmit(a.ctx, "err:cmdwait", err)
		}
		runtime.EventsEmit(a.ctx, "done")
	}()
	//	runtime.LogDebugf(a.ctx, "got %d bytes of data", len(stdoutStderr))
	//	runtime.LogDebugf(a.ctx, "%s", string(stdoutStderr))
	return `{"Action": "test", "Package": "` + p.Pkg + `"}`, nil
}
