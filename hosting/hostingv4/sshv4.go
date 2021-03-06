package hostingv4

import (
	"strconv"

	"github.com/PabloPie/go-gandi/hosting"
)

// sshkeyv4 represents an sshkey in v4, the only difference is the
// ID that is represented by an int in v4
type sshkeyv4 struct {
	Fingerprint string `xmlrpc:"fingerprint"`
	ID          int    `xmlrpc:"id"`
	Name        string `xmlrpc:"name"`
	Value       string `xmlrpc:"value"`
}

// CreateKey creates a key from the given name and value
func (h Hostingv4) CreateKey(name string, value string) (hosting.SSHKey, error) {
	params := []interface{}{
		map[string]string{
			"name":  name,
			"value": value,
		}}

	response := sshkeyv4{}
	err := h.Send("hosting.ssh.create", params, &response)
	if err != nil {
		return hosting.SSHKey{}, err
	}
	return h.keyFromID(response.ID), nil
}

// DeleteKey deletes an SSH Key
func (h Hostingv4) DeleteKey(key hosting.SSHKey) error {
	id, err := strconv.Atoi(key.ID)
	if err != nil {
		return err
	}
	params := []interface{}{id}
	var response = false
	err = h.Send("hosting.ssh.delete", params, &response)
	if response {
		return nil
	}
	return err
}

// KeyFromName returns the key with name `name`, or an empty
// object if the key doesn't exist
func (h Hostingv4) KeyFromName(name string) hosting.SSHKey {
	params := []interface{}{
		map[string]string{
			"name": name,
		}}
	response := []sshkeyv4{}
	err := h.Send("hosting.ssh.list", params, &response)
	if err != nil || len(response) < 1 {
		return hosting.SSHKey{}
	}
	return h.keyFromID(response[0].ID)
}

// ListKeys lists every available key, without the corresponding values
func (h Hostingv4) ListKeys() []hosting.SSHKey {
	response := []sshkeyv4{}
	_ = h.Send("hosting.ssh.list", []interface{}{}, &response)

	var keys = []hosting.SSHKey{}
	for _, key := range response {
		// Getting also the value of a key is optional...
		fullkey := h.keyFromID(key.ID)
		keys = append(keys, fullkey)
	}
	return keys
}

// Helper functions

// keyFromID is an internal function to get a general hosting.SSHKey from a v4 ID
func (h Hostingv4) keyFromID(id int) hosting.SSHKey {
	params := []interface{}{id}
	response := sshkeyv4{}
	_ = h.Send("hosting.ssh.info", params, &response)
	return toSSHKey(response)
}

// Conversion functions

// toSSHKey transforms a v4 hosting.SSHKey to a generic one
func toSSHKey(key sshkeyv4) hosting.SSHKey {
	return hosting.SSHKey{
		ID:          strconv.Itoa(key.ID),
		Fingerprint: key.Fingerprint,
		Name:        key.Name,
		Value:       key.Value,
	}
}
