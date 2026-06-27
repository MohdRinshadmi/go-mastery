package fixed

// Classify maps an exam score to a band.
//
// Contract:
//
//	score <  40          -> "fail"
//	40 <= score <  70    -> "pass"
//	score >= 70          -> "distinction"
//
// FIX: the distinction threshold is `score >= 70` (not `> 70`), so a
// score of exactly 70 is a distinction, matching the spec.
func Classify(score int) string {
	switch {
	case score < 40:
		return "fail"
	case score >= 70: // FIXED: was `score > 70`
		return "distinction"
	default:
		return "pass"
	}
}
