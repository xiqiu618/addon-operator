module github.com/flant/addon-operator

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/flant/shell-operator v1.0.2-0.20210524145531-147332b951de // branch: master
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-openapi/spec v0.19.3
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.7
	github.com/hashicorp/go-multierror v1.0.0
	github.com/kennygrant/sanitize v1.2.4
	github.com/onsi/gomega v1.9.0
	github.com/peterbourgon/mergemap v0.0.0-20130613134717-e21c03b7a721
	github.com/prometheus/client_golang v1.7.1
	github.com/segmentio/go-camelcase v0.0.0-20160726192923-7085f1e3c734
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.7.5
	github.com/tidwall/sjson v1.1.6
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/satori/go.uuid.v1 v1.2.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.11
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.19.11
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/go-openapi/validate => github.com/flant/go-openapi-validate v0.19.4-0.20200313141509-0c0fba4d39e1 // branch: fix_in_body_0_19_7

replace k8s.io/apimachinery => k8s.io/apimachinery v0.19.11
