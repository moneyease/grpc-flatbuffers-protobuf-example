package main

import (
	"testing"
)

func BenchmarkFlat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fbSendFile("lessthan20MBTest.doc")
	}
	//	b.ReportMetric(float64(compares)/float64(b.N), "compares/op")
}
func BenchmarkProto(b *testing.B) {
	for n := 0; n < b.N; n++ {
		pbSendFile("lessthan20MBTest.doc")
	}
	//	b.ReportMetric(float64(compares)/float64(b.N), "compares/op")
}
