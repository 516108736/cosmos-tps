package deploy

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Deploy struct {
	nameToIp   map[string]string
	SSHSession map[string]*SSHSession
}

func NewDeploy() *Deploy {
	d := Deploy{nameToIp: make(map[string]string), SSHSession: make(map[string]*SSHSession)}

	d.nameToIp["u1"] = "127.0.0.1"
	d.nameToIp["u2"] = "39.105.8.235"
	d.nameToIp["u3"] = "39.107.124.39"

	d.SSHSession["u1"] = NewSSHConnect("root", "123456", d.nameToIp["u1"], 22)
	d.SSHSession["u2"] = NewSSHConnect("root", "1783105689qQ", d.nameToIp["u2"], 22)
	d.SSHSession["u3"] = NewSSHConnect("root", "1783105689qQ", d.nameToIp["u3"], 22)
	return &d
}

func (d *Deploy) Stop() {
	for _, v := range d.SSHSession {
		v.RunCmd("kill -9 `ps -ef | grep gaiad | awk '{print $2}'`")
	}
}

func (d *Deploy) Ready() {
	for name, u := range d.SSHSession {
		u.RunCmd("rm -rf " + homePath)
		u.RunCmd("mkdir -p " + homePath)
		u.RunCmd("rm -rf /root/.gaia*")
		u.SendFile("./cdata.zip", homePath)
		u.RunCmd("unzip cdata.zip")
		fmt.Println("send cdata.zip end", name, d.nameToIp[name])
	}
}

var (
	homePath      = "/opt/g/"
	d             = flag.String("d", "", "send or query")
	localIP       = flag.String("ip", "", "ip")
	localPassword = flag.String("password", "", "ip")

	local = NewSSHConnect("root", "goquarkchain", "138.68.224.79", 22)
)

func (d *Deploy) SendAndGenLog() {
	local = NewSSHConnect("root", *localPassword, *localIP, 22)
	go local.RunCmd("cd /opt/g/ && chmod +x cosmos-tps && ./cosmos-tps -typ=send >> send.log 2>&1 &")
	time.Sleep(5 * time.Second)
	go local.RunCmd("cd /opt/g/ && chmod +x cosmos-tps && ./cosmos-tps -typ=query >> query.log 2>&1 &")
	time.Sleep(5 * time.Second)
	go local.RunCmd("cd /opt/g/ && chmod +x cosmos-tps && ./cosmos-tps -typ=deploy -d=cm >> cm.log 2>&1 &")
	time.Sleep(5 * time.Second)
	fmt.Println("end SendAndGenLog")
}

func (d *Deploy) MakeFile() {
	u1SSH := d.SSHSession["u1"]
	u1SSH.GetFile("/opt/g/log.log", "./")
	u1SSH.GetFile("/opt/g/send.log", "./")
	u1SSH.GetFile("/opt/g/query.log", "./")
	u1SSH.GetFile("/opt/g/cm.log", "./")
	local.RunCmd("cd /opt/cc/   &&  cat log.log send.log  query.log cm.log>> new.log")
}

func (d *Deploy) CpuAndMem() {
	local = NewSSHConnect("root", *localPassword, *localIP, 22)
	for true {
		cpu := local.RunCmdAndGetOutPut("top -bn1 | grep Cpu")
		mem := local.RunCmdAndGetOutPut("top -bn1 | grep \"KiB Mem\"")
		fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "ip", *localIP, "cpu-----", strings.Replace(cpu, "\n", "", -1), "mem-----", strings.Replace(mem, "\n", "", -1))

		time.Sleep(20 * time.Second)
	}

}

func (d *Deploy) init() {
	hh := "/opt/cc/"
	local.RunCmd("rm -rf " + hh)
	local.RunCmd("mkdir -p " + hh)
	local.SendFile("./cosmos-tps", hh)
	local.SendFile("./gaiad", hh)
	local.SendFile("./gaiacli", hh)
	local.SendFile("./scfTx1.json", hh)
	local.SendFile("./scfTx2.json", hh)
	local.SendFile("./scfTx3.json", hh)
	local.SendFile("./scfTx4.json", hh)

	local.SendFile("./k.zip", hh)
	local.SendFile("./config.toml", hh)
	local.SendFile("./genesis.json", hh)
}

func Start() {
	flag.Parse()
	dp := NewDeploy()

	switch *d {
	case "stop":
		dp.Stop()
	case "ready":
		dp.Stop()
		dp.Ready()
	case "cm":
		dp.CpuAndMem()
	case "init":
		dp.init()
	case "send":
		dp.SendAndGenLog()
	case "file":
		dp.MakeFile()
	}

}
