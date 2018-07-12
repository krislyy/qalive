package concat

import (
	"testing"
)

func TestConcatMP4ToFLVConcat(t *testing.T) {
	urls := []string{
		"/Users/kaolafm/Desktop/demo1.mp4",
		"/Users/kaolafm/Desktop/demo2.mp4",
		"/Users/kaolafm/Desktop/demo3.mp4",
		"/Users/kaolafm/Desktop/demo2.mp4",
		"/Users/kaolafm/Desktop/demo3.mp4",
		"/Users/kaolafm/Desktop/demo1.mp4",
	}

	target_url := "/Users/kaolafm/Desktop/demo1.flv"
	if err := ConcatMP4ToFLV(urls, target_url); err != nil {
		t.Error("concat error " , err)
	}
}