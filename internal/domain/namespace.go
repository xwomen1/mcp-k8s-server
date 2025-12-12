package domain

type Namespace string

const (
	NamespaceDefault Namespace = "default"
	NamespaceSystem  Namespace = "kube-system"
	NamespaceAll     Namespace = ""
)
