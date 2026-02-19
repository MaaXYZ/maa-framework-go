package main

import (
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4/tools/api-check/internal/checker"
)

func main() {
	os.Exit(checker.Run())
}
