<a name="unreleased"></a>
## [Unreleased]

### Features

- Put link to Tekton PipelineRun in Gerrit [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Update CodeReview Pipeline checkout for Gerrit VCS [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Put gerrit message for build pipeline [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Add Gerrit CodeReview/Build EventListeners [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Implement gerrit voting from pipelines [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Add GitHub integration [EPMDEDP-10429](https://jiraeu.epam.com/browse/EPMDEDP-10429)
- Create webhook in GitLab [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Add support for GitLab [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Add Build Pipeline with Gradle build tool [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Update Tasks and Pipelines [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Add EDP getversion step [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Implement Review/Build pipeline for Python Application [EPMDEDP-10432](https://jiraeu.epam.com/browse/EPMDEDP-10432)
- Use nexus from EDP as default python pypi mirror [EPMDEDP-10432](https://jiraeu.epam.com/browse/EPMDEDP-10432)
- Implement GO Build and Code Review Tekton pipelines [EPMDEDP-10433](https://jiraeu.epam.com/browse/EPMDEDP-10433)
- Enable gerrit support for JavaScript Application [EPMDEDP-10434](https://jiraeu.epam.com/browse/EPMDEDP-10434)
- Add java11 build pipeline [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Add default and edp versioning type into the pipelines [EPMDEDP-10506](https://jiraeu.epam.com/browse/EPMDEDP-10506)
- Add pruner for pipelineruns and taskruns [EPMDEDP-10509](https://jiraeu.epam.com/browse/EPMDEDP-10509)
- Add deployment templates for EDP Interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Implement EDP interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Implement github/gitlab events processing [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Generalize Review/Build Pipelines for Java11 [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Add commit-validate task [EPMDEDP-10588](https://jiraeu.epam.com/browse/EPMDEDP-10588)
- Implement helm-docs task [EPMDEDP-10592](https://jiraeu.epam.com/browse/EPMDEDP-10592)
- Build and Code-review pipelines for Dotnet [EPMDEDP-10593](https://jiraeu.epam.com/browse/EPMDEDP-10593)
- Add unit tests for interceptor [EPMDEDP-10599](https://jiraeu.epam.com/browse/EPMDEDP-10599)
- Implement HTTPS connection [EPMDEDP-10600](https://jiraeu.epam.com/browse/EPMDEDP-10600)

### Bug Fixes

- Fix EDP interceptor deployment configuration [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Fix deployment template mapping [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Add CODEBASE_NAME parameter to dotnet review [EPMDEDP-10589](https://jiraeu.epam.com/browse/EPMDEDP-10589)
- Fix port for interceptor service [EPMDEDP-10600](https://jiraeu.epam.com/browse/EPMDEDP-10600)

### Code Refactoring

- Move github-set-status task to tasks dir [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Move secrets under resource section [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Refactor GitHub WebHook creation [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Rename files for tekton-resources [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Create PVC during pipelineRun [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Service account edp-kinako might be already created [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Switch to nexus in specific namespace [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Update labels [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Introduce helm-chart as installer [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Remove specific namespace definition from resource [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Use generalized approach for go tool [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Replace jenkins user with edp-ci [EPMDEDP-10640](https://jiraeu.epam.com/browse/EPMDEDP-10640)
- Change helm charts structure [EPMDEDP-10645](https://jiraeu.epam.com/browse/EPMDEDP-10645)
- Use templates approach to generate resources [EPMDEDP-10649](https://jiraeu.epam.com/browse/EPMDEDP-10649)

### Routine

- Add github voting step to pipeline [EPMDEDP-10429](https://jiraeu.epam.com/browse/EPMDEDP-10429)
- Disable verbosity for gitlab-create-webhook task [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Add Hadolint and helm-lint tasks [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Add Code-review pipeline [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Add Java8 and Java11 version for the Maven pipelines [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Delete version-type parameter and get it from codebase cr [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Implement CodeReview/Build pipelines for JavaScript application [EPMDEDP-10434](https://jiraeu.epam.com/browse/EPMDEDP-10434)
- Add get-version for Gradle, Maven, Npm, Python, Go build tools [EPMDEDP-10506](https://jiraeu.epam.com/browse/EPMDEDP-10506)
- Add unittests for pipelines [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add unittests for helm templates [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add unittests for charts baseline [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add tests for pruner section in charts [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add Dockerfile for edp-interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)

### Documentation

- Update README file with general information [EPMDEDP-10649](https://jiraeu.epam.com/browse/EPMDEDP-10649)


<a name="v0.1.0"></a>
## v0.1.0 - 2022-08-23

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.1.0...HEAD
