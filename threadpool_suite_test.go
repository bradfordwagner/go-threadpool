package threadpool

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestThreadpool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Threadpool Suite")
}
