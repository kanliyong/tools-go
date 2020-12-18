package cmd

type Cluster struct {
	name       string
	tkeId      string
	kubeconfig string
	env        string
}

var ClusterArray = []*Cluster{
	&Cluster{
		name:       "pre-a",
		tkeId:      "cls-0dlxshdd",
		kubeconfig: "cls-0dlxshdd-config",
		env:        "proenv",
	},
	&Cluster{
		name:       "pre-b",
		tkeId:      "cls-j74pglw5",
		kubeconfig: "cls-j74pglw5-config",
		env:        "proenv-b",
	},
	&Cluster{
		name:       "release-b",
		tkeId:      "cls-azglyxmh",
		kubeconfig: "cls-azglyxmh-config",
		env:        "release-b",
	},
	&Cluster{
		name:       "release-d",
		tkeId:      "cls-cz6628et",
		kubeconfig: "cls-cz6628et-config",
		env:        "release-d",
	},
}
