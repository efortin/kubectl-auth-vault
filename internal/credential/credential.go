package credential

import (
	"encoding/json"
	"fmt"
	"io"
)

type ExecCredential struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Status     ExecCredentialStatus `json:"status"`
}

type ExecCredentialStatus struct {
	Token string `json:"token"`
}

func New(token string) *ExecCredential {
	return &ExecCredential{
		APIVersion: "client.authentication.k8s.io/v1",
		Kind:       "ExecCredential",
		Status: ExecCredentialStatus{
			Token: token,
		},
	}
}

func Output(w io.Writer, token string) error {
	cred := New(token)
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal ExecCredential: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func (e *ExecCredential) JSON() ([]byte, error) {
	return json.Marshal(e)
}
