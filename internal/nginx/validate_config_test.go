package nginx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestValidateConfigSuccess(t *testing.T) {
	successOutput := "nginx: the configuration file /etc/nginx/nginx.conf syntax is ok\n" +
		"nginx: configuration file /etc/nginx/nginx.conf test is successful"

	host := runtimehosttest.NewFakeHost().
		WithRunResult("nginx", []string{"-t"}, runtimehost.CommandResult{ExitCode: 0, Stderr: successOutput}, nil)
	manager := nginx.NewManager(host)

	result, err := manager.ValidateConfig(context.Background())
	if err != nil {
		t.Fatalf("ValidateConfig() error = %v", err)
	}
	if !result.Valid {
		t.Errorf("Valid = false, want true (output: %s)", result.Output)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}
}

func TestValidateConfigSyntaxError(t *testing.T) {
	failOutput := `nginx: [emerg] unexpected "}" in /etc/nginx/sites-enabled/example.conf:10
nginx: configuration file /etc/nginx/nginx.conf test failed`

	host := runtimehosttest.NewFakeHost().
		WithRunResult("nginx", []string{"-t"},
			runtimehost.CommandResult{ExitCode: 1, Stderr: failOutput},
			&runtimehost.ExecutionError{Command: "nginx", Args: []string{"-t"}, ExitCode: 1, Stderr: failOutput, Err: errors.New("exit status 1")},
		)
	manager := nginx.NewManager(host)

	result, err := manager.ValidateConfig(context.Background())
	if err != nil {
		t.Fatalf("ValidateConfig() error = %v, want nil (an invalid config is not a Go error)", err)
	}
	if result.Valid {
		t.Fatal("expected Valid = false for a broken config")
	}
	if len(result.Errors) != 1 || result.Errors[0] != `nginx: [emerg] unexpected "}" in /etc/nginx/sites-enabled/example.conf:10` {
		t.Errorf("Errors = %v, want the single [emerg] line", result.Errors)
	}
}

func TestValidateConfigNginxNotRunnable(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("nginx", []string{"-t"},
			runtimehost.CommandResult{ExitCode: -1},
			&runtimehost.ExecutionError{Command: "nginx", Args: []string{"-t"}, ExitCode: -1, Err: errors.New("exec: \"nginx\": executable file not found in $PATH")},
		)
	manager := nginx.NewManager(host)

	_, err := manager.ValidateConfig(context.Background())
	if err == nil {
		t.Fatal("expected an error when nginx never actually ran")
	}
}
