package cmd

import "testing"

func TestWorkerCmdDefaults(t *testing.T) {
	cmd := newWorkerCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	if got := cmd.Flags().Lookup("concurrency").DefValue; got != "1" {
		t.Errorf("default concurrency = %q, want 1", got)
	}
	if cmd.Flags().Lookup("queue-url") == nil || cmd.Flags().Lookup("queue") == nil {
		t.Fatal("worker command is missing RabbitMQ flags")
	}
}
