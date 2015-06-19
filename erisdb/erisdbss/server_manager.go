package erisdbss

import (
	"bufio"
	"fmt"
	"github.com/eris-ltd/eris-db/files"
	"github.com/eris-ltd/eris-db/server"
	"github.com/tendermint/tendermint/binary"
	. "github.com/tendermint/tendermint/common"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	REAPER_TIMEOUT   = 5 * time.Second
	REAPER_THRESHOLD = 10 * time.Second
	// Ports to new server processes are PORT_BASE + i, where i is an integer from the pool.
	PORT_BASE = 29000
	// How long are we willing to wait for a process to become ready.
	PROC_START_TIMEOUT = 3 * time.Second
	// Name of the process executable.
	EXECUTABLE_NAME = "erisdb"
)

// Executable processes. These are used to wrap processes that are .
type ExecProcess interface {
	Start(chan<- error)
	Kill() error
}

// Wrapper for exec.Cmd. Will wait for a token from stdout using line scanning.
type CmdProcess struct {
	cmd   *exec.Cmd
	token string
}

func newCmdProcess(cmd *exec.Cmd, token string) *CmdProcess {
	return &CmdProcess{cmd, token}
}

func (this *CmdProcess) Start(doneChan chan<- error) {
	log.Debug("Starting erisdb process")
	reader, errSP := this.cmd.StdoutPipe()

	if errSP != nil {
		doneChan <- errSP
	}

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	if errStart := this.cmd.Start(); errStart != nil {
		doneChan <- errStart
		return
	}
	fmt.Println("process started, waiting for token")
	for scanner.Scan() {
		text := scanner.Text()
		log.Debug(text)
		if strings.Index(text, this.token) != -1 {
			log.Debug("Token found", "token", this.token)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		doneChan <- fmt.Errorf("Error reading from process stdout:", err)
		return
	}
	log.Debug("ErisDB server ready.")
	doneChan <- nil
}

func (this *CmdProcess) Kill() error {
	return this.cmd.Process.Kill()
}

// A serve task. This wraps a running process. It was designed to run 'erisdb' processes.
type ServeTask struct {
	sp          ExecProcess
	workDir     string
	start       time.Time
	maxDuration time.Duration
	port        uint16
}

// Create a new serve task.
func newServeTask(port uint16, workDir string, maxDuration time.Duration, process ExecProcess) *ServeTask {
	return &ServeTask{
		process,
		workDir,
		time.Now(),
		maxDuration,
		port,
	}
}

// Catches events that callers subscribe to and adds them to an array ready to be polled.
type ServerManager struct {
	mtx      *sync.Mutex
	idPool   *server.IdPool
	maxProcs uint
	running  []*ServeTask
	reap     bool
	baseDir  string
}

//
func NewServerManager(maxProcs uint, baseDir string) *ServerManager {
	sm := &ServerManager{
		mtx:      &sync.Mutex{},
		idPool:   server.NewIdPool(maxProcs),
		maxProcs: maxProcs,
		running:  make([]*ServeTask, 0),
		reap:     true,
		baseDir:  baseDir,
	}
	go reap(sm)
	return sm
}

func reap(sm *ServerManager) {
	if !sm.reap {
		return
	}
	time.Sleep(REAPER_TIMEOUT)
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	// The processes are added in order so just read from bottom of array until
	// a time is below reaper threshold, then break.
	for len(sm.running) > 0 {
		task := sm.running[0]
		if task.maxDuration != 0 && time.Since(task.start) > task.maxDuration {
			log.Debug("[SERVER REAPER] Closing down server.", "port", task.port)
			task.sp.Kill()
			sm.running = sm.running[1:]
			sm.idPool.ReleaseId(uint(task.port - PORT_BASE))
		} else {
			break
		}
	}
	go reap(sm)
}

// Add a new erisdb process to the list.
func (this *ServerManager) add(data *RequestData) (*ResponseData, error) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	config := server.DefaultServerConfig()
	// Port is PORT_BASE + a value between 1 and the max number of servers.
	id, errId := this.idPool.GetId()
	if errId != nil {
		return nil, errId
	}
	port := uint16(PORT_BASE + id)
	config.Bind.Port = port

	folderName := fmt.Sprintf("testnode%d", port)
	workDir, errCWD := this.createWorkDir(data, config, folderName)
	if errCWD != nil {
		return nil, errCWD
	}

	// TODO ...

	// Create a new erisdb process.
	cmd := exec.Command(EXECUTABLE_NAME, workDir)
	proc := &CmdProcess{cmd, "DONTMINDME55891"}

	errSt := waitForProcStarted(proc)

	if errSt != nil {
		return nil, errSt
	}

	maxDur := time.Duration(data.MaxDuration) * time.Second
	if maxDur == 0 {
		maxDur = REAPER_THRESHOLD
	}

	st := newServeTask(port, workDir, maxDur, proc)
	this.running = append(this.running, st)

	// TODO add validation data. The node should ideally return some post-deploy state data
	// and send it back with the server URL, so that the validity of the chain can be
	// established client-side before starting the tests.
	return &ResponseData{fmt.Sprintf("%d", port)}, nil
}

// Add a new erisdb process to the list.
func (this *ServerManager) killAll() {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	for len(this.running) > 0 {
		task := this.running[0]
		log.Debug("Closing down server.", "port", task.port)
		task.sp.Kill()
		this.running = this.running[1:]
		this.idPool.ReleaseId(uint(task.port - PORT_BASE))
	}
}

// Creates a temp folder for the tendermint/erisdb node to run in.
// Folder name is port based, so port=1337 meens folder="testnode1337"
// Old folders are cleared out. before creating them, and the server will
// clean out all of this servers workdir (defaults to ~/.edbservers) when
// starting and when stopping.
func (this *ServerManager) createWorkDir(data *RequestData, config *server.ServerConfig, folderName string) (string, error) {

	workDir := path.Join(this.baseDir, folderName)
	os.RemoveAll(workDir)
	errED := EnsureDir(workDir)
	if errED != nil {
		return "", errED
	}

	cfgName := path.Join(workDir, "config.toml")
	scName := path.Join(workDir, "server_conf.toml")
	pvName := path.Join(workDir, "priv_validator.json")
	genesisName := path.Join(workDir, "genesis.json")

	// Write config.
	errCFG := files.WriteFileRW(cfgName, []byte(TendermintConfigDefault))
	if errCFG != nil {
		return "", errCFG
	}
	log.Info("File written.", "name", cfgName)

	// Write validator.
	errPV := writeJSON(pvName, data.PrivValidator)
	if errPV != nil {
		return "", errPV
	}
	log.Info("File written.", "name", pvName)

	// Write genesis
	errG := writeJSON(genesisName, data.Genesis)
	if errG != nil {
		return "", errG
	}
	log.Info("File written.", "name", genesisName)

	// Write server config.
	errWC := server.WriteServerConfig(scName, config)
	if errWC != nil {
		return "", errWC
	}
	log.Info("File written.", "name", scName)
	return workDir, nil
}

// Used to write json files using tendermints binary package.
func writeJSON(file string, v interface{}) error {
	var n int64
	var errW error
	fo, errC := os.Create(file)
	if errC != nil {
		return errC
	}
	binary.WriteJSON(v, fo, &n, &errW)
	if errW != nil {
		return errW
	}
	errL := fo.Close()
	if errL != nil {
		return errL
	}
	return nil
}

func waitForProcStarted(proc ExecProcess) error {
	timeoutChan := make(chan struct{})
	doneChan := make(chan error)
	done := new(bool)
	go func(b *bool) {
		time.Sleep(PROC_START_TIMEOUT)
		if !*b {
			timeoutChan <- struct{}{}
		}
	}(done)
	go proc.Start(doneChan)
	var errSt error
	select {
	case errD := <-doneChan:
		errSt = errD
		*done = true
		break
	case <-timeoutChan:
		_ = proc.Kill()
		errSt = fmt.Errorf("Process start timed out")
		break
	}
	return errSt
}
