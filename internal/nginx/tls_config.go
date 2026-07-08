package nginx

// TLSConfig names where a virtual host's certificate and key live. It is
// modeled now so VirtualHost has a stable place for it, but nothing in this
// iteration reads or acts on it — issuing or managing certificates
// (Certbot, Let's Encrypt) is out of scope until a later iteration.
type TLSConfig struct {
	Enabled         bool
	CertificatePath string
	PrivateKeyPath  string
}
