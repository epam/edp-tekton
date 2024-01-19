<a name="unreleased"></a>
## [Unreleased]

### Features

- Provision default TriggerTemplate for CD Pipeline ([#96](https://github.com/epam/edp-tekton/issues/96))
- Align promotion procedure to the new format ([#96](https://github.com/epam/edp-tekton/issues/96))
- Implement deploy pipeline based on Argo ApplicationSet ([#96](https://github.com/epam/edp-tekton/issues/96))
- Add custom pipelines for SAM tool ([#92](https://github.com/epam/edp-tekton/issues/92))

### Bug Fixes

- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))
- Change versioning for helm lib default([#101](https://github.com/epam/edp-tekton/issues/101))
- Add chart-dir parameter to helm pipeline([#101](https://github.com/epam/edp-tekton/issues/101))
- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))
- Failed push-to-jira step in build pipeline on okd ([#94](https://github.com/epam/edp-tekton/issues/94))
- Fix Service name for Ingress object of the EventListeners CR ([#93](https://github.com/epam/edp-tekton/issues/93))
- Align cache endpoint with service name ([#89](https://github.com/epam/edp-tekton/issues/89))
- Fix tekton cache service name ([#89](https://github.com/epam/edp-tekton/issues/89))

### Code Refactoring

- Simplify Git provider EventListeners and Ingress handling ([#93](https://github.com/epam/edp-tekton/issues/93))

### Routine

- Add dependencies between tasks([#103](https://github.com/epam/edp-tekton/issues/103))
- Bump tekton-cache version ([#102](https://github.com/epam/edp-tekton/issues/102))
- Bump tekton-cache helm chart version ([#102](https://github.com/epam/edp-tekton/issues/102))
- Update image on tekton-cache chart ([#102](https://github.com/epam/edp-tekton/issues/102))
- Add yaml files to semgrep ignore list ([#96](https://github.com/epam/edp-tekton/issues/96))
- Align naming convention for Argo CD task ([#96](https://github.com/epam/edp-tekton/issues/96))
- Bump sonarscanner image ([#99](https://github.com/epam/edp-tekton/issues/99))
- Switch deploy pipeline to tekton-cd-pipeline image ([#98](https://github.com/epam/edp-tekton/issues/98))
- Remove unused parameters([#97](https://github.com/epam/edp-tekton/issues/97))
- Align event-listener name for routes([#93](https://github.com/epam/edp-tekton/issues/93))
- Add cache for custom pipelines ([#95](https://github.com/epam/edp-tekton/issues/95))
- Bump golang.org/x/crypto from 0.14.0 to 0.17.0 ([#91](https://github.com/epam/edp-tekton/issues/91))
- Remove volume workspace from commit-validate task ([#78](https://github.com/epam/edp-tekton/issues/78))
- Remove volume workspace from getDefaultVersion task ([#78](https://github.com/epam/edp-tekton/issues/78))
- Remove volume workspace from getDefaultVersion task ([#78](https://github.com/epam/edp-tekton/issues/78))
- Update release steps subsequence ([#90](https://github.com/epam/edp-tekton/issues/90))
- Update current development version ([#89](https://github.com/epam/edp-tekton/issues/89))

### Documentation

- Update README md file ([#132](https://github.com/epam/edp-tekton/issues/132))


<a name="v0.10.2"></a>
## [v0.10.2] - 2024-01-19
### Bug Fixes

- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))
- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))


<a name="v0.10.1"></a>
## [v0.10.1] - 2023-12-18
### Bug Fixes

- Align cache endpoint with service name ([#89](https://github.com/epam/edp-tekton/issues/89))


<a name="v0.10.0"></a>
## [v0.10.0] - 2023-12-18
### Features

- Ensure commit-validate checks the lenght of the commit msg ([#87](https://github.com/epam/edp-tekton/issues/87))
- Make it possible to deploy cache chart with edp-tekton ([#74](https://github.com/epam/edp-tekton/issues/74))
- Implement cache in dotnet pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Implement cache in gradle pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Implement cache in JS pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Implement cache in python pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Add backstage custom pipelines ([#77](https://github.com/epam/edp-tekton/issues/77))
- Implement cache in maven pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Add tekton cache chart ([#83](https://github.com/epam/edp-tekton/issues/83))
- Automate rekor uuid in release tag ([#81](https://github.com/epam/edp-tekton/issues/81))
- Implement cache capabilities for golang pipelines ([#74](https://github.com/epam/edp-tekton/issues/74))
- Add ingress-enabled parameter for tekton event Listener with a check before creating the ingress resource ([#76](https://github.com/epam/edp-tekton/issues/76))
- Add ingress-enabled parameter for tekton dashboard with a check before creating the ingress resource ([#75](https://github.com/epam/edp-tekton/issues/75))
- Publish Dependency Track report for the specific version ([#71](https://github.com/epam/edp-tekton/issues/71))
- Add e2e tests for simple gerrit deployment ([#68](https://github.com/epam/edp-tekton/issues/68))

### Bug Fixes

- Update images for autotest sonar scan ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update images for autotest sonar scan ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update autotest tasks ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update dotnet push source path ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update dotnet push source path ([#74](https://github.com/epam/edp-tekton/issues/74))
- Make possible work with kaniko without region parameter ([#118](https://github.com/epam/edp-tekton/issues/118))
- We don't need source in workspaces subpath in fetch step ([#74](https://github.com/epam/edp-tekton/issues/74))
- Fix checkout to source sub-path ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update security task defenitions on go codebases ([#70](https://github.com/epam/edp-tekton/issues/70))
- Align edp-npm task to EDP repository ([#68](https://github.com/epam/edp-tekton/issues/68))

### Code Refactoring

- Return back to native python task ([#68](https://github.com/epam/edp-tekton/issues/68))
- Consolidate npm stages under single stage edp-npm ([#68](https://github.com/epam/edp-tekton/issues/68))
- Merge fastapi and flask frameworks into single template ([#68](https://github.com/epam/edp-tekton/issues/68))
- Rename python-edp to edp-python task ([#68](https://github.com/epam/edp-tekton/issues/68))
- Introduce EDP specific task for fastapi flow ([#68](https://github.com/epam/edp-tekton/issues/68))
- Introduce EDP specific task for general python flow ([#68](https://github.com/epam/edp-tekton/issues/68))
- Change gerrit notification approach ([#67](https://github.com/epam/edp-tekton/issues/67))
- Move gerrit related parts to separate file ([#67](https://github.com/epam/edp-tekton/issues/67))

### Testing

- Refactor e2e flow to reduce flaky tests ([#72](https://github.com/epam/edp-tekton/issues/72))
- Change resource creation flow ([#68](https://github.com/epam/edp-tekton/issues/68))
- Create pipelinerun to ensure Tekton stack is ready for testing ([#68](https://github.com/epam/edp-tekton/issues/68))
- Add github, gitlab cases to e2e tests ([#68](https://github.com/epam/edp-tekton/issues/68))
- Update chart dependencies for test installation ([#68](https://github.com/epam/edp-tekton/issues/68))

### Routine

- Update current development version ([#89](https://github.com/epam/edp-tekton/issues/89))
- Update current development version ([#89](https://github.com/epam/edp-tekton/issues/89))
- Update access right for npm-build task ([#74](https://github.com/epam/edp-tekton/issues/74))
- Align autotests pipeline params ([#88](https://github.com/epam/edp-tekton/issues/88))
- Merge steps of security task to reduce the number of containers([#87](https://github.com/epam/edp-tekton/issues/87))
- Merge the steps of the push-to-jira task into a single step to avoid the necessity of using volumes ([#87](https://github.com/epam/edp-tekton/issues/87))
- Make it possible to use external tekton cache ([#74](https://github.com/epam/edp-tekton/issues/74))
- Update tekton-autotest image version ([#74](https://github.com/epam/edp-tekton/issues/74))
- Add resources to tekton pruner ([#86](https://github.com/epam/edp-tekton/issues/86))
- Make SAST integration optional([#85](https://github.com/epam/edp-tekton/issues/85))
- Set parameter ctLint.validateMaintainer to false by default ([#84](https://github.com/epam/edp-tekton/issues/84))
- Update tekton cache tasks ([#74](https://github.com/epam/edp-tekton/issues/74))
- Ignore CodeQL scan for some files ([#74](https://github.com/epam/edp-tekton/issues/74))
- Ready for kind to be ready before starting deployment ([#74](https://github.com/epam/edp-tekton/issues/74))
- Apply new pruner approach ([#82](https://github.com/epam/edp-tekton/issues/82))
- Update custom edp images([#80](https://github.com/epam/edp-tekton/issues/80))
- Align terraform infrastructure and lib pipelines for tfenv usage ([#73](https://github.com/epam/edp-tekton/issues/73))
- Switch PipelineRun from v1beta1 to v1 ([#72](https://github.com/epam/edp-tekton/issues/72))
- Switch Task and Pipeline from v1beta1 to v1 ([#72](https://github.com/epam/edp-tekton/issues/72))
- Bump ct-lint version ([#69](https://github.com/epam/edp-tekton/issues/69))
- Relax resource requests for tekton tasks ([#67](https://github.com/epam/edp-tekton/issues/67))
- Remove deprecated pipelines ([#67](https://github.com/epam/edp-tekton/issues/67))
- Optimize custom pipelines flow by merging related tasks ([#67](https://github.com/epam/edp-tekton/issues/67))
- Use google analytics during docs build ([#65](https://github.com/epam/edp-tekton/issues/65))
- Update current development version ([#65](https://github.com/epam/edp-tekton/issues/65))


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

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.10.2...HEAD
[v0.10.2]: https://github.com/epam/edp-tekton/compare/v0.10.1...v0.10.2
[v0.10.1]: https://github.com/epam/edp-tekton/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/epam/edp-tekton/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/epam/edp-tekton/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/epam/edp-tekton/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/epam/edp-tekton/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/epam/edp-tekton/compare/v0.5.0...v0.6.0
