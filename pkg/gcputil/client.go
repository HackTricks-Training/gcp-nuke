package gcputil

import (
	"fmt"
	"os"

	"google.golang.org/api/option"
)

type GCPClient interface {
	Close() error
}

type Credentials struct {
	Keyfile string
	Project string
}

func (c *Credentials) UseAppDefaultCreds() bool {
	return c.Keyfile == ""
}

func (c *Credentials) Validate() error {
	if !c.UseAppDefaultCreds() {
		_, err := os.Stat(c.Keyfile)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Credentials are invalid: Key File '%s' not found", c.Keyfile)
			}
			return fmt.Errorf("Error validating credentials: %w", err)
		}
	}
	return nil
}

func (c *Credentials) GetNewClientOptions() (options []option.ClientOption) {
	options = []option.ClientOption{}
	if !c.UseAppDefaultCreds() {
		options = append(options, option.WithCredentialsFile(c.Keyfile))
	}
	return options
}
