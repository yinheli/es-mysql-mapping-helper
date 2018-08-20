package main

type Mappings struct {
	Settings *Settings           `json:"settings"`
	Mappings map[string]*Columns `json:"mappings"`
}

type Settings struct {
	Index *Index `json:"index"`
}

type Index struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

type Columns struct {
	Properties map[string]*Column `json:"properties"`
}

type Column struct {
	Type           string `json:"type"`
	Analyzer       string `json:"analyzer,omitempty"`
	SearchAnalyzer string `json:"search_analyzer,omitempty"`
}

type DbColumn struct {
	Name string
	Type string
}
