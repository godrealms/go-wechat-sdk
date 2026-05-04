package types

// PlatformCertificate represents one platform certificate
type PlatformCertificate struct {
	SerialNo           string              `json:"serial_no"`
	EffectiveTime      string              `json:"effective_time"`
	ExpireTime         string              `json:"expire_time"`
	EncryptCertificate *EncryptCertificate `json:"encrypt_certificate"`
}

// EncryptCertificate holds the encrypted certificate data
type EncryptCertificate struct {
	Algorithm      string `json:"algorithm"`       // AEAD_AES_256_GCM
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	CipherText     string `json:"ciphertext"`
}

// GetCertificatesResult is the result of GetCertificates
type GetCertificatesResult struct {
	Data []*PlatformCertificate `json:"data"`
}
