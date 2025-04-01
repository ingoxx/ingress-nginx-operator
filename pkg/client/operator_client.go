package client

import "sigs.k8s.io/controller-runtime/pkg/client"

type OperatorClientImp struct {
	Client client.Client
}

func NewOperatorClientImp(client client.Client) *OperatorClientImp {
	return &OperatorClientImp{Client: client}
}

func (c *OperatorClientImp) GetClient() client.Client {
	return c.Client
}
