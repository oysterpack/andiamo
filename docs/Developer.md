# Standards
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [12-Factor App](https://12factor.net/)

# Go Tools
- running tests with code coverage
  ```
  go test -covermode count -coverprofile cover.out
  # launches browser with coverage report
  go tool cover -html cover.out
  ```

# Code Commit Checklist
- [ ] go fmt
- [ ] go lint
- [ ] go vet
- [ ] unit tests
- [ ] benchmark tests

# Technology Stack
- Dependency Injection
  - [fx](https://github.com/uber-go/fx)
- Logging
  - [zerolog](https://github.com/rs/zerolog)
- Config
  - [envconfig](https://github.com/kelseyhightower/envconfig)
- Errors
