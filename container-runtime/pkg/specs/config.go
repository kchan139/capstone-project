package specs

type ContainerConfig struct {
    RootFS      RootfsConfig      `json:"root"`
    Process     ProcessConfig     `json:"process"`
    Hostname    string            `json:"hostname"`
    // Mounts      []MountConfig     `json:"mounts,omitempty"`
}

type RootfsConfig struct {
    Path     string `json:"path"`
    Readonly bool   `json:"readonly,omitempty"`
}

type ProcessConfig struct {
    Args        []string `json:"args"`
    Env         []string `json:"env,omitempty"`
    Cwd         string   `json:"cwd,omitempty"`
    Terminal    bool     `json:"terminal"`
    User        *User    `json:"user,omitempty"`
}

type User struct {
    UID            int   `json:"uid"`
    GID            int   `json:"gid"`
    AdditionalGids []int `json:"additionalGids,omitempty"`
}


// type MountConfig struct {
//     Source      string   `json:"source"`
//     Destination string   `json:"destination"`
//     Type        string   `json:"type,omitempty"`
//     Options     []string `json:"options,omitempty"`
// }