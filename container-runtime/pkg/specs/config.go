package specs

type ContainerConfig struct {
	RootFS      RootfsConfig  `json:"root"`
	Process     ProcessConfig `json:"process"`
	Hostname    string        `json:"hostname,omitempty"`
	ContainerId string        `json:"containerId,omitempty"`
	Linux       LinuxConfig   `json:"linux"`
}

type RootfsConfig struct {
	Path     string `json:"path"`
	Readonly bool   `json:"readonly,omitempty"`
}

type ProcessConfig struct {
	Args     []string `json:"args"`
	Env      []string `json:"env,omitempty"`
	Cwd      string   `json:"cwd,omitempty"`
	Terminal bool     `json:"terminal"`
	User     *User    `json:"user,omitempty"`
}

type User struct {
	UID            int   `json:"uid"`
	GID            int   `json:"gid"`
	AdditionalGids []int `json:"additionalGids,omitempty"`
}

type LinuxConfig struct {
	Resources *LinuxResources `json:"resources,omitempty"`
	Network   *LinuxNetwork   `json:"network,omitempty"`
}

type LinuxResources struct {
	CPU    *CPUConfig    `json:"cpu,omitempty"`
	Memory *MemoryConfig `json:"memory,omitempty"`
	Pids   *PidsConfig   `json:"pids,omitempty"`
}

type CPUConfig struct {
	Shares int64 `json:"shares,omitempty"`
	Quota  int64 `json:"quota,omitempty"`
	Period int64 `json:"period,omitempty"`
}
type MemoryConfig struct {
	Limit       int64 `json:"limit,omitempty"`
	Reservation int64 `json:"reservation,omitempty"`
	Swap        int64 `json:"swap,omitempty"`
}

type PidsConfig struct {
	Limit int64 `json:"limit,omitempty"`
}

type LinuxNetwork struct {
	EnableNetwork  bool     `json:"enableNetwork"`
	ContainerIP    string   `json:"containerIP"`
	GatewayIP      string   `json:"gatewayIP"`
	VethHost       string   `json:"vethHost"`
	VethContainer  string   `json:"vethContainer"`
	DNS            []string `json:"dns,omitempty"`
	FirewallScript string   `json:"firewallScript,omitempty"`
}
