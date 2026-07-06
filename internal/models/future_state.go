package models

import "time"

// The types in this file are not populated or read by any logic yet. They
// reserve the shape of data that process monitoring, deployment history and
// reverse-proxy/SSL management will need to persist, so the Application
// entity will not need to be reshaped when those features are built.

// RuntimeMetrics captures point-in-time operational data about a running
// application. Reserved for the process-management iteration.
type RuntimeMetrics struct {
	PID         int           `json:"pid,omitempty"`
	CPUPercent  float64       `json:"cpu_percent,omitempty"`
	MemoryBytes uint64        `json:"memory_bytes,omitempty"`
	Uptime      time.Duration `json:"uptime,omitempty"`
	LogFilePath string        `json:"log_file_path,omitempty"`
}

// DeploymentHistory records the outcome of the most recent deployment.
// Reserved for the deployment-pipeline iteration.
type DeploymentHistory struct {
	LastDeployedAt *time.Time `json:"last_deployed_at,omitempty"`
	LastCommitHash string     `json:"last_commit_hash,omitempty"`
}

// SSLCertificate records the TLS certificate bound to an application's
// domain. Reserved for the certificate-management iteration.
type SSLCertificate struct {
	Enabled   bool       `json:"enabled"`
	Issuer    string     `json:"issuer,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ReverseProxyConfig stores the generated reverse-proxy configuration for an
// application. Reserved for the Nginx/Apache management iteration.
type ReverseProxyConfig struct {
	Nginx  string `json:"nginx,omitempty"`
	Apache string `json:"apache,omitempty"`
}
