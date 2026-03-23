package build_info

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	tAssert "github.com/stretchr/testify/assert"
)

var assert *tAssert.Assertions

func TestBuildInfo(t *testing.T) {
	assert = tAssert.New(GinkgoT())
	RunSpecs(t, "BuildInfo Suite")
}
