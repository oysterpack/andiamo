# Standards
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [12-Factor App](https://12factor.net/)
- [How to write a Git commit message](https://chris.beams.io/posts/git-commit/)
- Uncle Bob martin

# Go Tools
- running tests with code coverage
  ```
  go test -covermode count -coverprofile cover.out
  # launches browser with coverage report
  go tool cover -html cover.out
  ```

# Code Commit Checklist
- [ ] go fmt
- [ ] [golint](https://github.com/golang/lint)
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
