package kinds

config: {
	kind: "Config"
        group: "app-o11y-config-manager"
	apiResource: {}
	codegen: {
		frontend: true
		backend: true
	}
	current: "v1"
	versions: {
		"v1": {
			schema: {
				spec: {
					enabled: bool
				}
			}
		}
	}
}

