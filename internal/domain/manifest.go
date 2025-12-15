package domain

type ApplyResult struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Action    string `json:"action"`
	DryRun    bool   `json:"dry_run"`
}
