package analyzer

import "time"

func SetTTL(r *RecordInfo) time.Duration {
	ttl, err := DetermineReason(r)
	if err != nil {
		return 0
	}

	return ttl
}
