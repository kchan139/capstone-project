package specs

import (
	"time"
	"os"
	"encoding/json"
)

// AuditEvent represents a single security event logged by the integrity monitor
type AuditEvent struct {
    Timestamp string           `json:"timestamp"`
    Type      string           `json:"type"`
    File      AuditFileInfo    `json:"file"`
    Process   AuditProcessInfo `json:"process"`
    Action    string           `json:"action"`
}

// AuditFileInfo contains metadata about the file involved in the event
type AuditFileInfo struct {
    Path        string `json:"path"`
    Permissions string `json:"permissions"`
    OwnerUID    int    `json:"owner_uid"`
    OwnerGID    int    `json:"owner_gid"`
}

// AuditProcessInfo contains metadata about the process that triggered the event
type AuditProcessInfo struct {
    PID int `json:"pid"`
    UID int `json:"uid"`
    GID int `json:"gid"`
}


func AuditAppendEvent(
	filepath string,
	eventType string,
	path string,
	pid int,
	procUID int,
	procGID int,
	fileUID int,
	fileGID int,
	filePerms os.FileMode,
	action string,
) error {

	e := AuditEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      eventType,
		Action:    action,
		File: AuditFileInfo{
			Path:        path,
			Permissions: filePerms.String(),
			OwnerUID:    fileUID,
			OwnerGID:    fileGID,
		},
		Process: AuditProcessInfo{
			PID: pid,
			UID: procUID,
			GID: procGID,
		},
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(e)
}