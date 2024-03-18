module github.com/ananasovich/app-o11y-config-manager

go 1.21

// Required as the project does not inherit the replace directive from grafana-app-sdk and grafana-plugin-sdk-go
replace github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.120.0
