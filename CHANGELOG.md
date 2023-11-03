<a name="unreleased"></a>
## [Unreleased]


<a name="v0.9.0"></a>
## [v0.9.0] - 2023-11-03
### Features

- Enable transparancy log upload to rekor fo release pipelines ([#64](https://github.com/epam/edp-tekton/issues/64))
- Enable dependency-track on sast task ([#59](https://github.com/epam/edp-tekton/issues/59))
- Upload transparency log to rekor for release pipelines ([#64](https://github.com/epam/edp-tekton/issues/64))
- Add e2e test to cd-pipeline-operator CI pipelines ([#61](https://github.com/epam/edp-tekton/issues/61))
- Enable resources for dashboard and eventlistener ([#54](https://github.com/epam/edp-tekton/issues/54))
- Implement integration with docker hub for openshift([#43](https://github.com/epam/edp-tekton/issues/43))
- Align helm-push-lib task to dockerhub integration ([#43](https://github.com/epam/edp-tekton/issues/43))
- Implement integration with docker hub ([#43](https://github.com/epam/edp-tekton/issues/43))

### Bug Fixes

- Change pipeline pattern for custom autotest codebase ([#49](https://github.com/epam/edp-tekton/issues/49))
- Versioning type to default edp-platform/common/autotests ([#52](https://github.com/epam/edp-tekton/issues/52))
- Make possible push chart with openshift registry ([#62](https://github.com/epam/edp-tekton/issues/62))
- Remove task dependency between sonar-cleanup and sonar ([#57](https://github.com/epam/edp-tekton/issues/57))
- Add parameter for helm-push-lib ([#47](https://github.com/epam/edp-tekton/issues/47))
- Update default versioning for dotnet app/lib ([#53](https://github.com/epam/edp-tekton/issues/53))
- Update pipelines for helm app/lib ([#51](https://github.com/epam/edp-tekton/issues/51))
- Fix parameter name ([#47](https://github.com/epam/edp-tekton/issues/47))
- Set image name pattern in kaniko task ([#47](https://github.com/epam/edp-tekton/issues/47))
- Add helm-push task for dockerhub integration ([#43](https://github.com/epam/edp-tekton/issues/43))
- Fix repository name pattern for js ([#43](https://github.com/epam/edp-tekton/issues/43))

### Routine

- Update current development version ([#65](https://github.com/epam/edp-tekton/issues/65))
- Rename push-report step in security task ([#59](https://github.com/epam/edp-tekton/issues/59))
- Rename SAST task to Security ([#59](https://github.com/epam/edp-tekton/issues/59))
- Migrate dep-track task to sast in custom pipelines ([#59](https://github.com/epam/edp-tekton/issues/59))
- Align edp autotest execution ([#60](https://github.com/epam/edp-tekton/issues/60))
- Migrate dep-track from rewiev to build custom-pipelines ([#59](https://github.com/epam/edp-tekton/issues/59))
- Update custom pipelines for new kaniko approach ([#47](https://github.com/epam/edp-tekton/issues/47))
- Optimize tekton tasks dependency ([#57](https://github.com/epam/edp-tekton/issues/57))
- Bump google.golang.org/grpc from 1.53.0 to 1.56.3 ([#58](https://github.com/epam/edp-tekton/issues/58))
- Upgrade pull request template ([#56](https://github.com/epam/edp-tekton/issues/56))
- Migrate edp-common from Jenkins to Tekton ([#52](https://github.com/epam/edp-tekton/issues/52))
- Add bing verification code ([#48](https://github.com/epam/edp-tekton/issues/48))
- Add robots.txt file ([#48](https://github.com/epam/edp-tekton/issues/48))
- Add indexnow verification ([#48](https://github.com/epam/edp-tekton/issues/48))
- Align helm tasks and pipelines for new config approach ([#47](https://github.com/epam/edp-tekton/issues/47))
- Update task dependencies for custom pipelines ([#47](https://github.com/epam/edp-tekton/issues/47))
- Migrate edp-autotests pipelines to Tekton ([#49](https://github.com/epam/edp-tekton/issues/49))
- Update pytest dependencies ([#50](https://github.com/epam/edp-tekton/issues/50))
- Bump golang.org/x/net from 0.9.0 to 0.17.0 ([#50](https://github.com/epam/edp-tekton/issues/50))
- Align cutom-pipeline for new kaniko approach ([#47](https://github.com/epam/edp-tekton/issues/47))
- Migrate platform pipelines to Tekton ([#48](https://github.com/epam/edp-tekton/issues/48))
- Align kaniko task for Openshift approach ([#47](https://github.com/epam/edp-tekton/issues/47))
- Change Kaniko parameter source ([#47](https://github.com/epam/edp-tekton/issues/47))
- Join Kaniko task for ECR and Harbor into one ([#47](https://github.com/epam/edp-tekton/issues/47))
- Join Kaniko task for Dockerhub and Harbor into one ([#47](https://github.com/epam/edp-tekton/issues/47))
- Bump sonar-scaner image ([#44](https://github.com/epam/edp-tekton/issues/44))
- Automate image bump in Chart.yaml for release process ([#42](https://github.com/epam/edp-tekton/issues/42))
- Add review and build pipelines for autotest type for java17 ([#40](https://github.com/epam/edp-tekton/issues/40))
- Use push to dockerhub instead of ecr-to-docker task in release pipelines ([#39](https://github.com/epam/edp-tekton/issues/39))
- Update current development version ([#41](https://github.com/epam/edp-tekton/issues/41))


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

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.9.0...HEAD
[v0.9.0]: https://github.com/epam/edp-tekton/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/epam/edp-tekton/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/epam/edp-tekton/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/epam/edp-tekton/compare/v0.5.0...v0.6.0
