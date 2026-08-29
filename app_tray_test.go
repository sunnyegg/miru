package main

import "testing"

func TestCloseDecision(t *testing.T) {
	tests := []struct {
		name       string
		keyPresent bool
		enabled    bool
		forceQuit  bool
		want       closeAction
	}{
		{name: "force quit", forceQuit: true, want: closeActionQuit},
		{name: "unset asks", want: closeActionAsk},
		{name: "disabled quits", keyPresent: true, want: closeActionQuit},
		{name: "enabled hides", keyPresent: true, enabled: true, want: closeActionHide},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := closeDecision(testCase.keyPresent, testCase.enabled, testCase.forceQuit)
			if got != testCase.want {
				t.Fatalf("closeDecision(%v, %v, %v) = %v, want %v",
					testCase.keyPresent, testCase.enabled, testCase.forceQuit, got, testCase.want)
			}
		})
	}
}
