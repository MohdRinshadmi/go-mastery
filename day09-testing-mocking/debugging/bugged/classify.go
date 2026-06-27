package bugged

// Classify maps an exam score to a band.
//
// The intended contract (from the spec):
//
//	score <  40          -> "fail"
//	40 <= score <  70    -> "pass"
//	score >= 70          -> "distinction"
//
// BUG: the distinction threshold is off by one. It uses `score > 70`
// instead of `score >= 70`, so a score of exactly 70 is wrongly
// classified as "pass". The table test in classify_test.go never
// catches this because it asserts against a hard-coded constant
// instead of the per-case `tc.want`.
func Classify(score int) string {
	switch {
	case score < 40:
		return "fail"
	case score > 70: // BUG: should be `score >= 70`
		return "distinction"
	default:
		return "pass"
	}
}
