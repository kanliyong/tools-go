package cmd

type Cluster struct {
	name       string
	tkeId      string
	kubeconfig string
	env        string
}

var ClusterArray = []*Cluster{
	&Cluster{
		name:       "test-a",
		tkeId:      "cls-2c5aysk1",
		kubeconfig: "cls-2c5aysk1-config",
		env:        "testenv",
	},
	&Cluster{
		name:       "test-b",
		tkeId:      "cls-9zg9knlr",
		kubeconfig: "cls-9zg9knlr-config",
		env:        "testenv-b",
	},
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
		name:       "release-c",
		tkeId:      "cls-0rxd0x77",
		kubeconfig: "cls-0rxd0x77-config",
		env:        "release-c",
	},
	&Cluster{
		name:       "release-d",
		tkeId:      "cls-cz6628et",
		kubeconfig: "cls-cz6628et-config",
		env:        "release-d",
	},
}
