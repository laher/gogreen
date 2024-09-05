package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

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
	lock    sync.RWMutex
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

func (a *App) GetState() (State, error) {
	// takes a write lock because it might update the state
	a.lock.Lock()
	defer a.lock.Unlock()
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

// Result represents a line of JSON output from `go test -json`
type Result struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Output  string `json:"Output"`
	Time    string `json:"Time"`
}

type Package struct {
	Pkg       string   `json:"pkg"`
	TestFuncs []string `json:"testFuncs"`
}

// TODO cwd
func (a *App) GetTestFuncs(p TestParams) ([]Package, error) {
	// takes a write lock
	a.lock.Lock()
	defer a.lock.Unlock()
	// any test func
	// (ignore example tests, benchmarks ...)
	cmd := exec.Command("go", "test", "-list=Test", "-json", p.Pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := []string{}
	lines = append(lines, strings.Split(string(out), "\n")...)
	lines = lines[:len(lines)-1]
	result := map[string]Package{}
	for _, line := range lines {
		r := Result{}
		err := json.Unmarshal([]byte(line), &r)
		if err != nil {
			return nil, err
		}
		if r.Action == "output" && strings.HasPrefix(r.Output, "Test") {
			tf := result[r.Package].TestFuncs
			result[r.Package] = Package{Pkg: r.Package, TestFuncs: append(tf, r.Output)}
		}
	}
	ret := []Package{}
	for _, r := range result {
		ret = append(ret, r)
	}
	return ret, nil
}

//go:embed frontend/src/assets/images
var images embed.FS

func (a *App) fsnotify(p TestParams) (*fsnotify.Watcher, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.state.Watching {
		return nil, errors.New("already watching")
	}
	if a.watcher != nil {
		panic("watcher should be nil when starting to watch")
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
					// closed channel
					return
				}
				if a.isRunning() {
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
					// closed channel
					return
				}
				runtime.LogErrorf(a.ctx, "watch error: %s", err)
			case <-a.stop:
				func() {
					a.lock.Lock()
					defer a.lock.Unlock()
					runtime.LogDebugf(a.ctx, "stop watching")
					a.watcher = nil
					a.state.Watching = false
				}()
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
	if !a.isWatching() {
		return errors.New("not watching")
	}
	a.stop <- struct{}{}
	return nil
}

func (a *App) isWatching() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.state.Watching
}

func (a *App) isRunning() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.state.Running
}

// Run returns a greeting for the given name
func (a *App) Run(p TestParams) (string, error) {
	if a.isRunning() {
		return "", errors.New("already running")
	}
	a.lock.Lock()
	defer a.lock.Unlock()
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

	emitOutput := func(name string, i io.ReadCloser) {
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
	go emitOutput("stdout", stdout)
	go emitOutput("stderr", stderr)
	go func() {
		// when run the tests - set 'not running' when done
		defer func() {
			a.lock.Lock()
			defer a.lock.Unlock()
			a.state.Running = false
		}()
		result := "PASS"
		err := cmd.Wait()
		if err != nil {
			result = "FAIL"
		}
		runtime.EventsEmit(a.ctx, "result", result)
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
