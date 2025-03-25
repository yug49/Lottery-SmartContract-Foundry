package config

type Workflows interface {
	Limits() WorkflowsLimits
}

type WorkflowsLimits interface {
	Global() int32
	PerOwner() int32
}
