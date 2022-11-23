package internal

import "fmt"

func ValidatePort(p int)  (int, error) {
	const minPort, maxPort = 1, 65535
	if p < minPort || p > maxPort {
		return p, fmt.Errorf("port %d out of range [%d:%d]", p, minPort, maxPort)
	}

	return p, nil
}
