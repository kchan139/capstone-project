package monitorfanotify

import (
	"fmt"
	"golang.org/x/sys/unix"
	"mrunc/pkg/specs"
)

// eventToFlag converts event string to fanotify flag
func eventToFlag(event string, isBlock bool) (uint64, error) {
	if isBlock {
		// Permission flags for blocking
		switch event {
		case "open":
			return unix.FAN_OPEN_PERM, nil
		case "read":
			return unix.FAN_ACCESS_PERM, nil
		case "exec":
			return unix.FAN_OPEN_EXEC_PERM, nil
		case "write":
			return unix.FAN_MODIFY, nil // Cannot intercept write, instead just log it
		default:
			return 0, fmt.Errorf("unknown event: %s", event)
		}
	} else {
		// Notification flags for audit
		switch event {
		case "open":
			return unix.FAN_OPEN, nil
		case "read":
			return unix.FAN_ACCESS, nil
		case "exec":
			return unix.FAN_OPEN_EXEC, nil
		case "write":
			return unix.FAN_MODIFY, nil
		default:
			return 0, fmt.Errorf("unknown event: %s", event)
		}
	}
}

// buildEventMask creates the combined event mask from watch rule
func BuildEventMask(rule specs.WatchRule) (uint64, error) {
	isBlock := rule.Action == "block"
	var mask uint64

	for _, event := range rule.Events {
		flag, err := eventToFlag(event, isBlock)
		if err != nil {
			return 0, err
		}
		mask |= flag
	}

	return mask, nil
}

// eventMaskToString converts event mask to readable string
func EventMaskToString(mask uint64) string {
	var events []string

	// Check notification flags
	if mask&unix.FAN_OPEN != 0 {
		events = append(events, "OPEN")
	}
	if mask&unix.FAN_ACCESS != 0 {
		events = append(events, "READ")
	}
	if mask&unix.FAN_MODIFY != 0 {
		events = append(events, "WRITE")
	}
	if mask&unix.FAN_OPEN_EXEC != 0 {
		events = append(events, "EXEC")
	}

	// Check permission flags
	if mask&unix.FAN_OPEN_PERM != 0 {
		events = append(events, "OPEN_PERM")
	}
	if mask&unix.FAN_ACCESS_PERM != 0 {
		events = append(events, "READ_PERM")
	}
	if mask&unix.FAN_OPEN_EXEC_PERM != 0 {
		events = append(events, "EXEC_PERM")
	}

	if len(events) == 0 {
		return "UNKNOWN"
	}

	result := events[0]
	for i := 1; i < len(events); i++ {
		result += "|" + events[i]
	}
	return result
}

// shouldBlockEvent determines if an event should be blocked based on rules
func shouldBlockEvent(path string, rules []specs.WatchRule) bool {
	for _, rule := range rules {
		// Check if path matches the rule's path or is under it
		if path == rule.Path || isUnderPath(path, rule.Path) {
			// If action is block, block this event
			if rule.Action == "block" {
				return true
			}
		}
	}
	return false
}

// isUnderPath checks if filePath is under dirPath
func isUnderPath(filePath, dirPath string) bool {
	// Simple prefix check - you might want to use filepath.HasPrefix or more robust checking
	return len(filePath) > len(dirPath) && filePath[:len(dirPath)] == dirPath
}

// eventMatchesMask checks if an event string matches the fanotify mask
func eventMatchesMask(event string, mask uint64) bool {
	switch event {
	case "open":
		return mask&(unix.FAN_OPEN|unix.FAN_OPEN_PERM) != 0
	case "read":
		return mask&(unix.FAN_ACCESS|unix.FAN_ACCESS_PERM) != 0
	case "exec":
		return mask&(unix.FAN_OPEN_EXEC|unix.FAN_OPEN_EXEC_PERM) != 0
	case "write":
		return mask&unix.FAN_MODIFY != 0
	default:
		return false
	}
}


// func ruleHasEvent(path string, eventType string, rules []WatchRule) bool {
//     for _, rule := range rules {
//         if path == rule.Path || isUnderPath(path, rule.Path) {
//             for _, event := range rule.Events {
//                 if event == eventType {
//                     return true
//                 }
//             }
//         }
//     }
//     return false
// }