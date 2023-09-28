<a name="unreleased"></a>
## [Unreleased]


<a name="v0.8.0"></a>
## [v0.8.0] - 2023-09-28
### Bug Fixes

- Update pattern for change version edp ([#36](https://github.com/epam/edp-tekton/issues/36))
- Update git-clone depth ([#20](https://github.com/epam/edp-tekton/issues/20))
- Update sonar variable definition ([#31](https://github.com/epam/edp-tekton/issues/31))
- Align naming for sonar_url parameter for java pipelines ([#31](https://github.com/epam/edp-tekton/issues/31))

### Routine

- Update current development version ([#41](https://github.com/epam/edp-tekton/issues/41))
- Align Tekton pipelines diff ([#37](https://github.com/epam/edp-tekton/issues/37))
- Implement signed image functionality during image push to Harbor ([#35](https://github.com/epam/edp-tekton/issues/35))
- Upgrade Go to 1.20 ([#34](https://github.com/epam/edp-tekton/issues/34))
- Update CHANGELOG.md ([#33](https://github.com/epam/edp-tekton/issues/33))
- Update current development version ([#33](https://github.com/epam/edp-tekton/issues/33))


<a name="v0.7.0"></a>
## [v0.7.0] - 2023-09-21
### Features

- Add EDP release pipelines for tekton ([#20](https://github.com/epam/edp-tekton/issues/20))
- Implement dependency track task for custom pipelines ([#16](https://github.com/epam/edp-tekton/issues/16))

### Bug Fixes

- Align naming for sonar_url parameter for java pipelines ([#31](https://github.com/epam/edp-tekton/issues/31))
- Run java8 sonar-scanner on runner with java11 autotests ([#31](https://github.com/epam/edp-tekton/issues/31))
- Run java8 sonar-scanner on runner with java11 maven ([#31](https://github.com/epam/edp-tekton/issues/31))
- Run java8 sonar-scanner on runner with java11 ([#31](https://github.com/epam/edp-tekton/issues/31))
- Add workspace to update-build-number tasks ([#30](https://github.com/epam/edp-tekton/issues/30))
- Sonar url for github lib ([#24](https://github.com/epam/edp-tekton/issues/24))
- Update parameters in helm pipelines ([#30](https://github.com/epam/edp-tekton/issues/30))
- Add sonar url to maven cm ([#23](https://github.com/epam/edp-tekton/issues/23))
- GitServer skipWebhookSSLVerification parameter ([#26](https://github.com/epam/edp-tekton/issues/26))
- Fix logic for python default versioning ([#74](https://github.com/epam/edp-tekton/issues/74))
- Remove NuGet token from output log ([#22](https://github.com/epam/edp-tekton/issues/22))
- Fix the execution sequence of update-build-number and sast tasks of NPM ([#17](https://github.com/epam/edp-tekton/issues/17))
- Refactor autotest-maven pipeline for GitHub VCS([#18](https://github.com/epam/edp-tekton/issues/18))
- Fix the execution sequence of update-build-number and sast tasks of Python ([#17](https://github.com/epam/edp-tekton/issues/17))
- Fix the execution sequence of update-build-number and sast tasks of Csharp ([#17](https://github.com/epam/edp-tekton/issues/17))
- Fix the execution sequence of update-build-number and sast tasks of Java ([#17](https://github.com/epam/edp-tekton/issues/17))
- Refactor autotest-maven pipeline ([#18](https://github.com/epam/edp-tekton/issues/18))

### Code Refactoring

- Align VCS secret name pattern ([#27](https://github.com/epam/edp-tekton/issues/27))
- Use helm Release Namespace instead of edpName value ([#25](https://github.com/epam/edp-tekton/issues/25))

### Routine

- Align release versions ([#33](https://github.com/epam/edp-tekton/issues/33))
- Use github as a default gitserver ([#32](https://github.com/epam/edp-tekton/issues/32))
- Deploy Tekton Dashboard with write permissions by default ([#28](https://github.com/epam/edp-tekton/issues/28))
- Update default gitlab server ([#29](https://github.com/epam/edp-tekton/issues/29))
- Update external component logic ([#24](https://github.com/epam/edp-tekton/issues/24))
- Align logic for default versioning ([#74](https://github.com/epam/edp-tekton/issues/74))
- Align sonar-operator pipelines ([#23](https://github.com/epam/edp-tekton/issues/23))
- Add test to tekton pipeline for sonar-operator ([#21](https://github.com/epam/edp-tekton/issues/21))
- Update container image for helm related tasks ([#19](https://github.com/epam/edp-tekton/issues/19))
- Update current development version ([#15](https://github.com/epam/edp-tekton/issues/15))

### Documentation

- Bump tekton version ([#59](https://github.com/epam/edp-tekton/issues/59))


<a name="v0.6.0"></a>
## [v0.6.0] - 2023-08-18

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.8.0...HEAD
[v0.8.0]: https://github.com/epam/edp-tekton/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/epam/edp-tekton/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/epam/edp-tekton/compare/v0.5.0...v0.6.0
