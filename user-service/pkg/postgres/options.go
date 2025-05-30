package postgres

import "fmt"

func (m Config) genConnectURL() string {
	auth := m.Username
	if m.Password != "" {
		auth = fmt.Sprintf("%s:%s", m.Username, m.Password)
	}

	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
		auth,
		m.Host,
		m.Port,
		m.Database,
		m.SSLMode)
}
