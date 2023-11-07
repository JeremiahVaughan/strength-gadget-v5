package model

type Deployment struct {
	Environment    string
	DesiredVersion string
	EnvVars        map[string]string
}
