package runtime

import (
	"os"
	"runtime"
	"time"
)

type OS struct {
	start time.Time
}

func NewOS() *OS {
	return &OS{
		start: time.Now(),
	}
}

func (o *OS) Tmpdir() string {
	return os.TempDir()
}

func (o *OS) Hostname() (string, error) {
	return os.Hostname()
}

func (o *OS) Uptime() float64 {
	return time.Now().Sub(o.start).Seconds()
}

func (o *OS) Platform() string {
	return runtime.GOOS
}

func (o *OS) Arch() string {
	return runtime.GOARCH
}

//os.endianness()
//os.type()
//os.release()
//os.loadavg()
//os.totalmem()
//os.freemem()
//os.cpus()
//os.networkInterfaces()
//os.EOL

func (o *OS) NumCPU() int {
	return runtime.NumCPU()
}

func (o *OS) NumGoroutine() int {
	return runtime.NumGoroutine()
}

func (o *OS) MemStats() *runtime.MemStats {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	return m
}
