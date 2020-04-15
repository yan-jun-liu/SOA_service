# SOA_service
Exercise: http://users.jyu.fi/~miselico/teaching/2015-2016/TIES456/advanced/task1/

# Prerequirement
1. Install standard-version:
   `npm i -g standard-version` More info:https://github.com/conventional-changelog/standard-version
2. Install conform for git commit message check:
   https://github.com/talos-systems/conform

# Build image
Run `make image`

# Run
Run `make run`


# test
Under folder `common` run `go test` to view the result
Under root folder run `go test -cover ./...` to view the test coverage