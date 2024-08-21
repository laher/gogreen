package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	// state to communicate to the frontend
	state State

	// internal state
	ctx     context.Context // wails requires a context.Context
	watcher *fsnotify.Watcher
	stop    chan struct{}
}

// state to communicate to the frontend
type State struct {
	CWD        string     `json:"cwd"`
	PkgList    []string   `json:"pkg_list"`
	Watching   bool       `json:"watching"`
	Running    bool       `json:"running"`
	TestParams TestParams `json:"test_params"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		stop: make(chan struct{}),
		state: State{
			CWD:        ".",
			TestParams: TestParams{Pkg: ".", Verbose: true},
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.state.CWD == "" {
		a.state.CWD = "." // simplest thing to do
	}
	err := os.Chdir(a.state.CWD)
	if err != nil {
		panic("cant change directory to " + a.state.CWD + ": " + err.Error())
	}
}

type TestParams struct {
	Pkg     string `json:"pkg"`
	Verbose bool   `json:"verbose"`
	Race    bool   `json:"race"`
	Run     string `json:"run"`
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

func (a *App) List() (State, error) {
	cmd := exec.Command("go", "list", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return State{}, err
	}
	lines := []string{".", "./..."}
	lines = append(lines, strings.Split(string(out), "\n")...)
	lines = lines[:len(lines)-1]

	a.state.PkgList = lines

	return a.state, nil
}

//go:embed frontend/src/assets/images
var images embed.FS

func (a *App) fsnotify(p TestParams) (*fsnotify.Watcher, error) {
	if a.state.Watching || a.watcher != nil {
		return nil, errors.New("already watching")
	}
	var err error
	a.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	a.watcher.Add(".")
	a.state.Watching = true

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				if a.state.Running {
					// debounce? ... in sime cases maybe set a flag saying we should start again right away?
					runtime.LogDebugf(a.ctx, "ignore watch event (already running tests): %+v", event)
					continue
				}
				runtime.LogDebugf(a.ctx, "watch event: %+v", event)
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
				a.state.Watching = false
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
	if !a.state.Watching {
		return errors.New("not watching")
	}
	a.stop <- struct{}{}
	return nil
}

// Run returns a greeting for the given name
func (a *App) Run(p TestParams) (string, error) {
	a.state.Running = true
	// count=1 to avoid caching
	params := []string{"test", "-json", "-count=1"}
	if p.Race {
		params = append(params, "-race")
	}
	if p.Verbose {
		params = append(params, "-v")
	}
	if p.Run != "" {
		params = append(params, "-run", p.Run)
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
		a.state.Running = false
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
		defer func() { a.state.Running = false }()
		err := cmd.Wait()
		result := "PASS"
		if err != nil {
			result = "FAIL"
		}
		runtime.EventsEmit(a.ctx, "result."+result, err)
		runtime.LogDebugf(a.ctx, "'go test' exited: %v", err)

		fw, err := os.Create("/tmp/logo.png")
		if err != nil {
			runtime.LogWarningf(a.ctx, "err creating image file: %s", err)
		} else {
			fr, err := images.Open("frontend/src/assets/images/logo-universal.png")
			if err != nil {
				runtime.LogWarningf(a.ctx, "err loading image data: %s", err)
			} else {
				_, err = io.Copy(fw, fr)
				if err != nil {
					runtime.LogWarningf(a.ctx, "err writing image data: %s", err)
				}
				fw.Close()
				fr.Close()
			}
		}

		err = beeep.Notify(
			fmt.Sprintf("test result - %s", result),
			"test finished. test "+result+"ED",
			"/tmp/logo.png")
		if err != nil {
			runtime.LogWarningf(a.ctx, "err sending notification: %s", err)
		}
		stdout.Close()
		stderr.Close()
	}()
	//	runtime.LogDebugf(a.ctx, "got %d bytes of data", len(stdoutStderr))
	//	runtime.LogDebugf(a.ctx, "%s", string(stdoutStderr))
	return `{"Action": "test", "Package": "` + p.Pkg + `"}`, nil
}
