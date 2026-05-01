package provenance

type ResourceDescriptor struct {
	Uri    string            `json:"uri"`
	Digest map[string]string `json:"digest"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

type BuilderParams struct {
	Packages        []string `json:"packages"`
	NoDeps          bool     `json:"no_deps,omitempty"`
	SourceDateEpoch int64    `json:"SOURCE_DATE_EPOCH,omitempty"`
}

type ExternalParameters struct {
}

type InternalParameters struct {
	BuilderParams *BuilderParams `json:"geschenkerbauer,omitempty"`
}

type BuildDefinition struct {
	BuildType            string                `json:"buildType"`
	ExternalParameters   *ExternalParameters   `json:"externalParameters"`
	InternalParameters   *InternalParameters   `json:"internalParameters"`
	ResolvedDependencies []*ResourceDescriptor `json:"resolvedDependencies"`
}

type Builder struct {
	Id                  string                `json:"id"`
	BuilderDependencies []*ResourceDescriptor `json:"builderDependencies,omitempty"`
	Version             map[string]string     `json:"version,omitempty"`
}

type BuildMetadata struct {
	InvocationID string `json:"invocationId"`
	StartedOn    string `json:"startedOn"`
	FinishedOn   string `json:"finishedOn"`
}

type RunDetails struct {
	Builder    *Builder              `json:"builder"`
	Metadata   *BuildMetadata        `json:"metadata,omitempty"`
	Byproducts []*ResourceDescriptor `json:"byproducts,omitempty"`
}

type Provenance struct {
	BuildDefinition *BuildDefinition `json:"buildDefinition"`
	RunDetails      *RunDetails      `json:"runDetails"`
}

type Statement struct {
	Type          string      `json:"_type"`
	Subject       []*Subject  `json:"subject"`
	PredicateType string      `json:"predicateType"`
	Predicate     *Provenance `json:"predicate"`
}
