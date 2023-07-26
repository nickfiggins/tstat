module github.com/nickfiggins/tstat

go 1.20

require (
	github.com/google/go-cmp v0.5.8
	github.com/stretchr/testify v1.8.2
	golang.org/x/exp v0.0.0-20230510235704-dd950f8aeaea
	golang.org/x/tools v0.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract [v0.0.1, v0.0.6] // From initial project that is now archived.
