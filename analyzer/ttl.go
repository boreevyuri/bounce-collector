package analyzer

func SetTTL(r *RecordInfo) int {
	ttl, err := DetermineReason(r)
	if err != nil {
		return 0
	}
	return ttl
}
