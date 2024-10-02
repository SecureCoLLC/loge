package loge

import (
	"fmt"
	"testing"
)

func TestLogName(t *testing.T) {
	fmt.Println("Testing getLogName")
	fmt.Println(getLogName("./logs"))
}
