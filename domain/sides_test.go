package domain_test

import (
	"testing"

	"github.com/ehpalumbo/go-diff/domain"
)

func TestParseDiffSideLeft(t *testing.T) {
	actual, err := domain.ParseDiffSide("left")
	if err != nil {
		t.Errorf("left side not recognized, got: %v", err.Error())
	}
	if actual != domain.LeftSide {
		t.Error("left side is not left")
	}
}

func TestParseDiffSideRight(t *testing.T) {
	actual, err := domain.ParseDiffSide("right")
	if err != nil {
		t.Errorf("right side not recognized, got: %v", err.Error())
	}
	if actual != domain.RightSide {
		t.Error("right side is not right")
	}
}

func TestParseDiffSideEmpty(t *testing.T) {
	_, err := domain.ParseDiffSide("")
	if err == nil {
		t.Error("empty side was accepted")
	}
	actualMessage := err.Error()
	if actualMessage != "invalid side value" {
		t.Errorf("empty side error is wrong, got: %v", actualMessage)
	}
}
