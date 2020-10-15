package process

import (
	"os"
	"time"

	"github.com/iamalsaher/interactor/pkg/pty"
)

//Process contains all the info about the Process
type Process struct {
	details *Details
	Pipe    *Pipes
	pty     *pty.PTY
	proc    *os.Process
	io      bool
	PID     int
}

//Start is used to finally Start the process
func (p *Process) Start(i *Interactor) (e error) {

	if p.io && i != nil {
		go i.Function(i.Input, i.Output)
	}

	if p.pty != nil {
		p.proc, e = newConPTYProcess(p.details.path, p.details.args, p.details.rundir, p.details.env, &p.pty.SIX)
	} else {
		p.proc, e = os.StartProcess(p.details.path, append([]string{p.details.path}, p.details.args...), &os.ProcAttr{
			Dir:   p.details.rundir,
			Env:   p.details.env,
			Sys:   nil,
			Files: []*os.File{p.Pipe.stdinR, p.Pipe.stdoutW, p.Pipe.stderrW},
		})
	}

	if e == nil {
		p.PID = p.proc.Pid

		//Setup timeout function
		if p.details.timeout > 0 {
			time.AfterFunc(p.details.timeout, func() {
				p.Kill()
			})
		}
	}
	return e
}

/*
ConnectIO is used to connect the input output handles
It attempts PTY as a primary mechanism else falls back to Pipes
If forcePTY is set then function errors out if pty cannot be aquired
*/
func (p *Process) ConnectIO(forcePTY bool) error {

	p.Pipe = new(Pipes)

	if pty, err := pty.NewPTY(); err == nil {
		p.pty = pty
		p.Pipe.StdinW = pty.Input
		p.Pipe.StdoutR = pty.Output
		p.Pipe.StderrR = pty.Output
		p.io = true
		p.Pipe.closer = append(p.Pipe.closer, pty.Input, pty.Output)
		return nil
	} else if forcePTY {
		return err
	}
	return setPipeIO(p)
}

//Kill is a wrapper around os.Process.Kill()
func (p *Process) Kill() error {
	if p.pty != nil {
		p.pty.Close()
	}
	return p.proc.Kill()
}
