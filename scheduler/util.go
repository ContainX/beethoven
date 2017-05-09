package scheduler

import "os"

func tlsEnabled(tlsCert, tlsCaCert, tlsKey string) bool {
	for _, v := range []string{tlsCert, tlsCaCert, tlsKey} {
		if e, err := pathExists(v); e && err == nil {
			return true
		}
	}
	return false
}

// pathExists returns whether the given file or directory exists or not
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
