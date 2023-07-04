<a name="unreleased"></a>
## [Unreleased]

### Features

- Add vulnerability image scanner task [EPMDEDP-11763](https://jiraeu.epam.com/browse/EPMDEDP-11763)
- Integration with public SonarQube [EPMDEDP-12074](https://jiraeu.epam.com/browse/EPMDEDP-12074)
- Add Next.js appliation [EPMDEDP-12080](https://jiraeu.epam.com/browse/EPMDEDP-12080)
- Implement helm-docs tekton task [EPMDEDP-12122](https://jiraeu.epam.com/browse/EPMDEDP-12122)
- Add Helm library pipelines [EPMDEDP-12135](https://jiraeu.epam.com/browse/EPMDEDP-12135)
- Add the ability to push container images to either ECR or Harbor registry [EPMDEDP-12181](https://jiraeu.epam.com/browse/EPMDEDP-12181)

### Bug Fixes

- Change helm-lint directory for edp-tekton review [EPMDEDP-11763](https://jiraeu.epam.com/browse/EPMDEDP-11763)
- Remove git-hooks for git-cli task [EPMDEDP-12113](https://jiraeu.epam.com/browse/EPMDEDP-12113)
- Align autotest runner for all vcs [EPMDEDP-12147](https://jiraeu.epam.com/browse/EPMDEDP-12147)
- Update pipeline run for autotest runner [EPMDEDP-12147](https://jiraeu.epam.com/browse/EPMDEDP-12147)
- Add step dependency [EPMDEDP-12161](https://jiraeu.epam.com/browse/EPMDEDP-12161)
- Fix helm-docs step on helm-helm-app pipeline [EPMDEDP-12176](https://jiraeu.epam.com/browse/EPMDEDP-12176)

### Routine

- Change kaniko-build to dockerbuild-verify in custom review pipelines [EPMDEDP-11763](https://jiraeu.epam.com/browse/EPMDEDP-11763)
- Update current development version [EPMDEDP-11826](https://jiraeu.epam.com/browse/EPMDEDP-11826)
- Align argocd tekton diff [EPMDEDP-12102](https://jiraeu.epam.com/browse/EPMDEDP-12102)
- Bump tekton-dashboard version to 0.36.1 [EPMDEDP-12106](https://jiraeu.epam.com/browse/EPMDEDP-12106)
- Enable SAST scan for build pipelines [EPMDEDP-12119](https://jiraeu.epam.com/browse/EPMDEDP-12119)
- Use default ct lint configs from root folder [EPMDEDP-12122](https://jiraeu.epam.com/browse/EPMDEDP-12122)
- Update the Kaniko Task to use its with Harbor registry [EPMDEDP-12181](https://jiraeu.epam.com/browse/EPMDEDP-12181)
- Implement trivy scan on sast step [EPMDEDP-12183](https://jiraeu.epam.com/browse/EPMDEDP-12183)
- Implement maven framework support for autotest [EPMDEDP-12189](https://jiraeu.epam.com/browse/EPMDEDP-12189)
- Bump alpine docker image to 3.18.2 [EPMDEDP-12253](https://jiraeu.epam.com/browse/EPMDEDP-12253)


<a name="v0.5.0"></a>
## [v0.5.0] - 2023-05-25
### Features

- Add pipelines for Helm application [EPMDEDP-11478](https://jiraeu.epam.com/browse/EPMDEDP-11478)
- Implement autotests in tekton [EPMDEDP-11660](https://jiraeu.epam.com/browse/EPMDEDP-11660)
- Add new JS frameworks [EPMDEDP-11760](https://jiraeu.epam.com/browse/EPMDEDP-11760)
- Add trivy-scan task to check image vulnerabilities [EPMDEDP-11763](https://jiraeu.epam.com/browse/EPMDEDP-11763)
- Add Gin Go framework [EPMDEDP-11836](https://jiraeu.epam.com/browse/EPMDEDP-11836)
- Add pipeline for dotnet 3.1 aplication [EPMDEDP-11881](https://jiraeu.epam.com/browse/EPMDEDP-11881)
- Add review pipeline re-triggering by comment /recheck [EPMDEDP-11899](https://jiraeu.epam.com/browse/EPMDEDP-11899)
- Enable support for maven multimodule project [EPMDEDP-11937](https://jiraeu.epam.com/browse/EPMDEDP-11937)
- Implement autotest as quality gate [EPMDEDP-11966](https://jiraeu.epam.com/browse/EPMDEDP-11966)
- Add terraform infrastructure pipeline [EPMDEDP-8292](https://jiraeu.epam.com/browse/EPMDEDP-8292)

### Bug Fixes

- Fix the sequence of stages run in the gerrit-kaniko-other-app-review pipeline [EPMDEDP-11763](https://jiraeu.epam.com/browse/EPMDEDP-11763)
- Fix Tekton dashboard link for pipelines [EPMDEDP-11835](https://jiraeu.epam.com/browse/EPMDEDP-11835)
- Fix nexus proxy usage for npm [EPMDEDP-11837](https://jiraeu.epam.com/browse/EPMDEDP-11837)
- Refactor helm-lint task [EPMDEDP-11873](https://jiraeu.epam.com/browse/EPMDEDP-11873)
- Refactor helm-lint task for Tekton helm pipelines [EPMDEDP-11873](https://jiraeu.epam.com/browse/EPMDEDP-11873)
- Fix ct-lint values [EPMDEDP-11909](https://jiraeu.epam.com/browse/EPMDEDP-11909)
- Add build step to review pipelines [EPMDEDP-11927](https://jiraeu.epam.com/browse/EPMDEDP-11927)
- Update configmap name in deployment [EPMDEDP-11966](https://jiraeu.epam.com/browse/EPMDEDP-11966)
- Wrong tag value for Helm app EDP strategy [EPMDEDP-11973](https://jiraeu.epam.com/browse/EPMDEDP-11973)
- Separate promotion task for autotests and cdpipeline [EPMDEDP-12005](https://jiraeu.epam.com/browse/EPMDEDP-12005)
- Moved the task to common and refactor to run pre-commit on OKD [EPMDEDP-12079](https://jiraeu.epam.com/browse/EPMDEDP-12079)
- Change permission to run pre-commit on OKD [EPMDEDP-12079](https://jiraeu.epam.com/browse/EPMDEDP-12079)
- Fix terraform infrastructure pipeline [EPMDEDP-8292](https://jiraeu.epam.com/browse/EPMDEDP-8292)

### Code Refactoring

- Align labels for PipelineRuns CR [EPMDEDP-12004](https://jiraeu.epam.com/browse/EPMDEDP-12004)

### Routine

- Update current development version [EPMDEDP-11472](https://jiraeu.epam.com/browse/EPMDEDP-11472)
- Remove go and npm cache to increase performance [EPMDEDP-11472](https://jiraeu.epam.com/browse/EPMDEDP-11472)
- Remove make test duplication for keycloak-operator [EPMDEDP-11552](https://jiraeu.epam.com/browse/EPMDEDP-11552)
- Refactor the `helm-lint` Tekton Task [EPMDEDP-11695](https://jiraeu.epam.com/browse/EPMDEDP-11695)
- Add ability to change tekton workspace size via helm values [EPMDEDP-11704](https://jiraeu.epam.com/browse/EPMDEDP-11704)
- Add unit tests for JS frameworks [EPMDEDP-11760](https://jiraeu.epam.com/browse/EPMDEDP-11760)
- Added task wait-for to create a wait queue for the task helm-push-gh-pages [EPMDEDP-11765](https://jiraeu.epam.com/browse/EPMDEDP-11765)
- Align argocd tekton diff [EPMDEDP-11766](https://jiraeu.epam.com/browse/EPMDEDP-11766)
- Remove .npmignore from npm task [EPMDEDP-11821](https://jiraeu.epam.com/browse/EPMDEDP-11821)
- Bump version to 0.5.0 [EPMDEDP-11826](https://jiraeu.epam.com/browse/EPMDEDP-11826)
- Update go version for GH Actions [EPMDEDP-11899](https://jiraeu.epam.com/browse/EPMDEDP-11899)
- Switch CodebaseImageStream tags creation to RFC3339 time format [EPMDEDP-11903](https://jiraeu.epam.com/browse/EPMDEDP-11903)
- Migrate ct-lint config to values file [EPMDEDP-11909](https://jiraeu.epam.com/browse/EPMDEDP-11909)
- Add templates for github issues [EPMDEDP-11928](https://jiraeu.epam.com/browse/EPMDEDP-11928)
- Bump tekton-dashboard version [EPMDEDP-11929](https://jiraeu.epam.com/browse/EPMDEDP-11929)
- Bump semgrep version [EPMDEDP-11949](https://jiraeu.epam.com/browse/EPMDEDP-11949)
- Update autotests [EPMDEDP-12004](https://jiraeu.epam.com/browse/EPMDEDP-12004)
- Update base image in autotests [EPMDEDP-12004](https://jiraeu.epam.com/browse/EPMDEDP-12004)
- Add labels to autotest pipeline [EPMDEDP-12017](https://jiraeu.epam.com/browse/EPMDEDP-12017)
- Update image version in tekton pipeline [EPMDEDP-12020](https://jiraeu.epam.com/browse/EPMDEDP-12020)
- Update image version in tekton pipeline [EPMDEDP-8292](https://jiraeu.epam.com/browse/EPMDEDP-8292)


<a name="v0.4.0"></a>
## [v0.4.0] - 2023-03-26
### Features

- Merge build and review TriggerTempletes into one for each [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Add pipelines for containers app [EPMDEDP-11320](https://jiraeu.epam.com/browse/EPMDEDP-11320)
- FastApi and Flask Tekton pipelines for Gerrit VCS [EPMDEDP-11359](https://jiraeu.epam.com/browse/EPMDEDP-11359)
- Add FastApi and Flask Tekton pipelines for Gitlab VCS [EPMDEDP-11359](https://jiraeu.epam.com/browse/EPMDEDP-11359)
- Add  FastApi and Flask Tekton pipelines for Github VCS [EPMDEDP-11359](https://jiraeu.epam.com/browse/EPMDEDP-11359)
- Add ability for kaniko-build to operate with openshift internal registry [EPMDEDP-11429](https://jiraeu.epam.com/browse/EPMDEDP-11429)
- Add RoleBinding to enable tekton build with openshift internal registry [EPMDEDP-11429](https://jiraeu.epam.com/browse/EPMDEDP-11429)
- Send notification to MSteams chat when Build pipeline failed [EPMDEDP-11464](https://jiraeu.epam.com/browse/EPMDEDP-11464)
- Implement read-only mode in tekton-dashboard [EPMDEDP-11467](https://jiraeu.epam.com/browse/EPMDEDP-11467)
- Add e2e Tekton Task [EPMDEDP-11483](https://jiraeu.epam.com/browse/EPMDEDP-11483)
- Implement Java 17 support for Gradle and Maven (Gerrit) [EPMDEDP-11558](https://jiraeu.epam.com/browse/EPMDEDP-11558)
- Implement Java 17 support for Gradle and Maven (Github) [EPMDEDP-11558](https://jiraeu.epam.com/browse/EPMDEDP-11558)
- Implement Java 17 support for Gradle and Maven (Gitlab) [EPMDEDP-11558](https://jiraeu.epam.com/browse/EPMDEDP-11558)
- Enable Route for gitlab/github el-listener on OKD [EPMDEDP-11587](https://jiraeu.epam.com/browse/EPMDEDP-11587)
- Implement Helm Chart support (Gerrit) [EPMDEDP-11599](https://jiraeu.epam.com/browse/EPMDEDP-11599)
- Implement Helm Chart support (Gitlab) [EPMDEDP-11599](https://jiraeu.epam.com/browse/EPMDEDP-11599)
- Implement Helm Chart support (Github) [EPMDEDP-11599](https://jiraeu.epam.com/browse/EPMDEDP-11599)
- Add task push chart to ecr [EPMDEDP-11599](https://jiraeu.epam.com/browse/EPMDEDP-11599)
- Implement  pipelines for .Net 6.0 framework for GitHub VCS [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)

### Bug Fixes

- Add JIRA_SERVER parameter to Tekton build pipelines for headlamp application [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Fix permissions error in npm-push on OKD [EPMDEDP-11329](https://jiraeu.epam.com/browse/EPMDEDP-11329)
- Align for common-library approach [EPMDEDP-11359](https://jiraeu.epam.com/browse/EPMDEDP-11359)
- Add updating helm chart dependencies in Makefile [EPMDEDP-11359](https://jiraeu.epam.com/browse/EPMDEDP-11359)
- Customization mkdocs Task [EPMDEDP-11432](https://jiraeu.epam.com/browse/EPMDEDP-11432)
- Fix the helm-lint Task for edp-tekton app [EPMDEDP-11443](https://jiraeu.epam.com/browse/EPMDEDP-11443)
- Bump npm version in headlamp pipeline [EPMDEDP-11458](https://jiraeu.epam.com/browse/EPMDEDP-11458)
- Align bash approach to push multiple helm packages to git in a loop [EPMDEDP-11509](https://jiraeu.epam.com/browse/EPMDEDP-11509)
- Tekton pipeline fails on OKD cluster for fastapi and flask applications [EPMDEDP-11609](https://jiraeu.epam.com/browse/EPMDEDP-11609)
- Change get-version-go-default task [EPMDEDP-11689](https://jiraeu.epam.com/browse/EPMDEDP-11689)
- Stop triggering the pipeline if the branch is not integrated with EDP [EPMDEDP-11714](https://jiraeu.epam.com/browse/EPMDEDP-11714)
- Add permission to interceptor to get codebasebranches [EPMDEDP-11714](https://jiraeu.epam.com/browse/EPMDEDP-11714)
- Add `HOME` environment variable into helm tasks [EPMDEDP-11751](https://jiraeu.epam.com/browse/EPMDEDP-11751)
- Specify targetPort for event-listener [EPMDEDP-11752](https://jiraeu.epam.com/browse/EPMDEDP-11752)

### Code Refactoring

- Align commit-validate Task [EPMDEDP-11288](https://jiraeu.epam.com/browse/EPMDEDP-11288)
- Bump mcdocs image version [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Bump tekton-dashboard version [EPMDEDP-11403](https://jiraeu.epam.com/browse/EPMDEDP-11403)
- Update headlamp review and build pipeline [EPMDEDP-11458](https://jiraeu.epam.com/browse/EPMDEDP-11458)

### Testing

- Unit tests for custom-pipelines [EPMDEDP-11291](https://jiraeu.epam.com/browse/EPMDEDP-11291)
- Unit tests for custom-pipelines [EPMDEDP-11291](https://jiraeu.epam.com/browse/EPMDEDP-11291)

### Routine

- Tekton deployment diff common-library [EPMDEDP-11218](https://jiraeu.epam.com/browse/EPMDEDP-11218)
- Tekton deployment diff [EPMDEDP-11218](https://jiraeu.epam.com/browse/EPMDEDP-11218)
- Update semgrep [EPMDEDP-11219](https://jiraeu.epam.com/browse/EPMDEDP-11219)
- Restore common-library [EPMDEDP-11265](https://jiraeu.epam.com/browse/EPMDEDP-11265)
- Align current development version [EPMDEDP-11265](https://jiraeu.epam.com/browse/EPMDEDP-11265)
- Remove edpName from values [EPMDEDP-11265](https://jiraeu.epam.com/browse/EPMDEDP-11265)
- Remove tekton common-library [EPMDEDP-11265](https://jiraeu.epam.com/browse/EPMDEDP-11265)
- Add custom pipelines for edp-tekton [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Add custom pipelines for edp-install [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Add custom pipelines for edp-admin-console [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Update Tekton build and review pipelines for headlamp application [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Add custom pipelines for edp-component-operator [EPMDEDP-11307](https://jiraeu.epam.com/browse/EPMDEDP-11307)
- Rename the pipelines for kaniko buildtool [EPMDEDP-11320](https://jiraeu.epam.com/browse/EPMDEDP-11320)
- Update custom tekton-dotnet image [EPMDEDP-11328](https://jiraeu.epam.com/browse/EPMDEDP-11328)
- Add request-limit resource block to Tasks [EPMDEDP-11374](https://jiraeu.epam.com/browse/EPMDEDP-11374)
- Implement tags for Tekton container [EPMDEDP-11385](https://jiraeu.epam.com/browse/EPMDEDP-11385)
- Use stable image tags for Tekton agents [EPMDEDP-11385](https://jiraeu.epam.com/browse/EPMDEDP-11385)
- Update pipeline-library dependencies [EPMDEDP-11403](https://jiraeu.epam.com/browse/EPMDEDP-11403)
- Add NPM-cache volume for npm tasks [EPMDEDP-11460](https://jiraeu.epam.com/browse/EPMDEDP-11460)
- Send notification to msteams chat when Build pipeline is failed [EPMDEDP-11464](https://jiraeu.epam.com/browse/EPMDEDP-11464)
- Change helm-docs task to golang taskRef in tekton custom-pipelines [EPMDEDP-11481](https://jiraeu.epam.com/browse/EPMDEDP-11481)
- Add Tekton Task for integration test [EPMDEDP-11483](https://jiraeu.epam.com/browse/EPMDEDP-11483)
- Create GitLab/GitHub EDPComponents [EPMDEDP-11516](https://jiraeu.epam.com/browse/EPMDEDP-11516)
- Update msteams message format for failed pipeline notification [EPMDEDP-11517](https://jiraeu.epam.com/browse/EPMDEDP-11517)
- Update git-chglog for edp-tekton [EPMDEDP-11518](https://jiraeu.epam.com/browse/EPMDEDP-11518)
- Bump golang.org/x/net from 0.5.0 to 0.8.0 [EPMDEDP-11578](https://jiraeu.epam.com/browse/EPMDEDP-11578)
- Remove make api-docs duplication [EPMDEDP-11599](https://jiraeu.epam.com/browse/EPMDEDP-11599)
- Add values for tekton common-library [EPMDEDP-11600](https://jiraeu.epam.com/browse/EPMDEDP-11600)
- Align image parameter to all pipelines [EPMDEDP-11601](https://jiraeu.epam.com/browse/EPMDEDP-11601)
- Make roleArn field optional [EPMDEDP-11604](https://jiraeu.epam.com/browse/EPMDEDP-11604)
- Bump Go version for all pipelines [EPMDEDP-11612](https://jiraeu.epam.com/browse/EPMDEDP-11612)
- Bump Go version for all pipelines [EPMDEDP-11612](https://jiraeu.epam.com/browse/EPMDEDP-11612)
- Update images for Tekton maven pipelines [EPMDEDP-11643](https://jiraeu.epam.com/browse/EPMDEDP-11643)
- Bump tekton-helm image version from epamedp/tekton-helm:0.1.3 to epamedp/tekton-helm:0.1.4 [EPMDEDP-11692](https://jiraeu.epam.com/browse/EPMDEDP-11692)
- Move the `e2e` task from build to review pipeline [EPMDEDP-11697](https://jiraeu.epam.com/browse/EPMDEDP-11697)
- Implement pipelines for .NET 6.0 (Gerrit) [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)
- Update pipelines for .Net 6.0 [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)
- Use container image from epamedp docker hub account [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)
- Implement library build and review pipelines for .NET 6.0 (Gerrit) [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)
- Implement library build and review pipelines for .NET 6.0 (Gitlab) [EPMDEDP-11731](https://jiraeu.epam.com/browse/EPMDEDP-11731)
- Implement pipelines for .NET 6.0 (Gitlab) [EPMDEDP-11740](https://jiraeu.epam.com/browse/EPMDEDP-11740)


<a name="v0.3.1"></a>
## [v0.3.1] - 2023-01-13
### Features

- Implement Kaniko cache [EPMDEDP-10532](https://jiraeu.epam.com/browse/EPMDEDP-10532)
- Implement Kaniko cache [EPMDEDP-10532](https://jiraeu.epam.com/browse/EPMDEDP-10532)
- Add commit-validate step to custom-pipelines [EPMDEDP-11210](https://jiraeu.epam.com/browse/EPMDEDP-11210)
- Add commit-validate step to common-library [EPMDEDP-11210](https://jiraeu.epam.com/browse/EPMDEDP-11210)
- Add commit-validate step to common-library [EPMDEDP-11210](https://jiraeu.epam.com/browse/EPMDEDP-11210)
- Enable Route for Tekton dashboard [EPMDEDP-11226](https://jiraeu.epam.com/browse/EPMDEDP-11226)

### Bug Fixes

- Allow default storageClass [EPMDEDP-11230](https://jiraeu.epam.com/browse/EPMDEDP-11230)
- Remove build version from sem version of Jira [EPMDEDP-11287](https://jiraeu.epam.com/browse/EPMDEDP-11287)

### Routine

- Convert application version to lowercase [EPMDEDP-11216](https://jiraeu.epam.com/browse/EPMDEDP-11216)
- Bump Semgrep version to 1.2.1, add --jobs flag to the executable command [EPMDEDP-11219](https://jiraeu.epam.com/browse/EPMDEDP-11219)
- Align Tekton dependencies [EPMDEDP-11226](https://jiraeu.epam.com/browse/EPMDEDP-11226)
- Change codebase parameter to codebasebranch in PipelineRun name [EPMDEDP-11293](https://jiraeu.epam.com/browse/EPMDEDP-11293)


<a name="v0.3.0"></a>
## [v0.3.0] - 2022-12-19

<a name="v0.2.9"></a>
## [v0.2.9] - 2022-12-17
### Bug Fixes

- Remove setting volume mount for gradle autotests [EPMDEDP-11217](https://jiraeu.epam.com/browse/EPMDEDP-11217)
- Fix build pipeline for maven java11 [EPMDEDP-11217](https://jiraeu.epam.com/browse/EPMDEDP-11217)


<a name="v0.2.8"></a>
## [v0.2.8] - 2022-12-16
### Bug Fixes

- Fix go-sdk build pipeline for gitlab [EPMDEDP-11075](https://jiraeu.epam.com/browse/EPMDEDP-11075)


<a name="v0.2.7"></a>
## [v0.2.7] - 2022-12-16
### Bug Fixes

- Align codereview for go-operator-sdk for GitHub [EPMDEDP-11075](https://jiraeu.epam.com/browse/EPMDEDP-11075)

### Routine

- Add interceptors to Role of tekton dashboard [EPMDEDP-11169](https://jiraeu.epam.com/browse/EPMDEDP-11169)


<a name="v0.2.6"></a>
## [v0.2.6] - 2022-12-16
### Bug Fixes

- Remove extra parameter from pipeline [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)


<a name="v0.2.5"></a>
## [v0.2.5] - 2022-12-16
### Bug Fixes

- Remove extra parameter from pipeline [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)


<a name="v0.2.4"></a>
## [v0.2.4] - 2022-12-15
### Routine

- Set default field for parameters [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)


<a name="v0.2.3"></a>
## [v0.2.3] - 2022-12-15
### Features

- Add Jira Task in each build pipeline [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)
- Return empty string when JiraServer is not defined [EPMDEDP-11190](https://jiraeu.epam.com/browse/EPMDEDP-11190)

### Routine

- Add finally-block which contains push-to-jira Task [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)
- Bump interceptor version to 0.2.3 [EPMDEDP-11186](https://jiraeu.epam.com/browse/EPMDEDP-11186)
- Bump tekton-dashboard version to 0.31.0 [EPMDEDP-11197](https://jiraeu.epam.com/browse/EPMDEDP-11197)


<a name="v0.2.2"></a>
## [v0.2.2] - 2022-12-13
### Features

- Update link to tekton-dashboard [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)
- Implement tekton-dashboard dependency [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)
- Implement tekton dashboard with impersonation [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)
- Implement Pipeline for Operator SDK [EPMDEDP-11075](https://jiraeu.epam.com/browse/EPMDEDP-11075)
- Use Interceptor instead of ClusterInterceptor [EPMDEDP-11138](https://jiraeu.epam.com/browse/EPMDEDP-11138)

### Bug Fixes

- Change include for review [EPMDEDP-11075](https://jiraeu.epam.com/browse/EPMDEDP-11075)
- Change Image Version for Operator Sdk [EPMDEDP-11075](https://jiraeu.epam.com/browse/EPMDEDP-11075)
- Rename golang-build to golang for gitlab pipelines [EPMDEDP-11144](https://jiraeu.epam.com/browse/EPMDEDP-11144)
- Remove changeNumber and patchsetNumber parameters from build pipelines for GitHub [EPMDEDP-11153](https://jiraeu.epam.com/browse/EPMDEDP-11153)

### Code Refactoring

- Remove unused tektonUrl param [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)
- Switch tekton dashboard from cluster to namespace [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)

### Routine

- Updated tekton dependencies [EPMDEDP-11027](https://jiraeu.epam.com/browse/EPMDEDP-11027)
- Use namespaced edp interceptor for EventListeners [EPMDEDP-11028](https://jiraeu.epam.com/browse/EPMDEDP-11028)
- Use Role instead of ClusterRole for interceptor [EPMDEDP-11028](https://jiraeu.epam.com/browse/EPMDEDP-11028)
- Bump up Kaniko to latest stable version [EPMDEDP-11088](https://jiraeu.epam.com/browse/EPMDEDP-11088)
- Update pruner logic [EPMDEDP-11109](https://jiraeu.epam.com/browse/EPMDEDP-11109)
- Define git-refspec parameter [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)
- Remove changeNumber and patchsetNumber from gitlab build pipelines [EPMDEDP-11143](https://jiraeu.epam.com/browse/EPMDEDP-11143)
- Add finally block which contains update-cbb Tasks [EPMDEDP-11183](https://jiraeu.epam.com/browse/EPMDEDP-11183)
- Add update-cbb block to common-library [EPMDEDP-11183](https://jiraeu.epam.com/browse/EPMDEDP-11183)


<a name="v0.2.1"></a>
## [v0.2.1] - 2022-12-05
### Bug Fixes

- GitHub event change target branch from head.ref to base.ref [EPMDEDP-11124](https://jiraeu.epam.com/browse/EPMDEDP-11124)

### Code Refactoring

- Use different hostnames for Github/GitLab EL [EPMDEDP-11078](https://jiraeu.epam.com/browse/EPMDEDP-11078)
- Change eventTypes from push to pull_request for triggering build pipelines [EPMDEDP-11124](https://jiraeu.epam.com/browse/EPMDEDP-11124)

### Routine

- Align values.yaml [EPMDEDP-10642](https://jiraeu.epam.com/browse/EPMDEDP-10642)
- Bump tekton version [EPMDEDP-10642](https://jiraeu.epam.com/browse/EPMDEDP-10642)
- Align the sequence of tasks for GitHub and GitLab [EPMDEDP-11124](https://jiraeu.epam.com/browse/EPMDEDP-11124)


<a name="v0.2.0"></a>
## [v0.2.0] - 2022-12-02
### Bug Fixes

- Define codebasebranch parameter for github flow [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)

### Code Refactoring

- Define gitProvider parameter [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)
- Rename secret for GitHub/GitLab [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)
- Update secret usage approach [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)

### Routine

- Get gerrit sshPort from global section [EPMDEDP-10642](https://jiraeu.epam.com/browse/EPMDEDP-10642)


<a name="v0.1.9"></a>
## [v0.1.9] - 2022-12-01
### Features

- Add gitlab, github gitservers provisioning [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)

### Bug Fixes

- Allow Kaniko to work with branches with slash [EPMDEDP-11067](https://jiraeu.epam.com/browse/EPMDEDP-11067)

### Code Refactoring

- Trigger pipelines for all Github branches [EPMDEDP-11077](https://jiraeu.epam.com/browse/EPMDEDP-11077)
- Explicitly define secret Keys for GitHub/GitLab [EPMDEDP-11119](https://jiraeu.epam.com/browse/EPMDEDP-11119)

### Routine

- Remove the cm settings block from TriggerTemplates update autotests and docs [EPMDEDP-11051](https://jiraeu.epam.com/browse/EPMDEDP-11051)
- Use common-library for GitHub and GitLab pipelines [EPMDEDP-11059](https://jiraeu.epam.com/browse/EPMDEDP-11059)
- Bump dependency version to 0.1.8 [EPMDEDP-11059](https://jiraeu.epam.com/browse/EPMDEDP-11059)


<a name="v0.1.8"></a>
## [v0.1.8] - 2022-11-30
### Features

- Implement CDPipeline in Tekton [EPMDEDP-11043](https://jiraeu.epam.com/browse/EPMDEDP-11043)
- Add EDP labels to PipelineRuns [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)
- Add volume for caching files of Go [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)

### Bug Fixes

- Add CODEBASEBRANCH_NAME parameter for gitlab java lib [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)

### Code Refactoring

- Add label for deploy Pipeline [EPMDEDP-11041](https://jiraeu.epam.com/browse/EPMDEDP-11041)

### Routine

- Enable Jira integration [EPMDEDP-11008](https://jiraeu.epam.com/browse/EPMDEDP-11008)
- Remove the cm settings block from TriggerTemplates get-version for gradle/maven [EPMDEDP-11051](https://jiraeu.epam.com/browse/EPMDEDP-11051)
- Remove the cm settings block from TriggerTemplates [EPMDEDP-11051](https://jiraeu.epam.com/browse/EPMDEDP-11051)
- Move common tasks of GitHub and GitLab pipelines to common-library [EPMDEDP-11059](https://jiraeu.epam.com/browse/EPMDEDP-11059)
- If storageClass is not specified, use default storageClass for go-cache volume [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)
- Remove gerrit-go-other-app-build-default Task and trigger folder [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)
- Align Tasks to a general form [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)
- Change go proxy link to an internal one [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)
- Align the sequence of tasks in the Review pipeline [EPMDEDP-11082](https://jiraeu.epam.com/browse/EPMDEDP-11082)


<a name="v0.1.7"></a>
## [v0.1.7] - 2022-11-26
### Features

- Grab codebase name from events [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)

### Bug Fixes

- Use codebase from interceptor [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)

### Code Refactoring

- Use codebase from interceptor [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)

### Routine

- Align for work with other framework [EPMDEDP-11008](https://jiraeu.epam.com/browse/EPMDEDP-11008)


<a name="v0.1.6"></a>
## [v0.1.6] - 2022-11-24
### Features

- Put codebasebranch name into interceptor payload [EPMDEDP-11057](https://jiraeu.epam.com/browse/EPMDEDP-11057)
- Return codebase name as a part of EDP interceptor payload [EPMDEDP-11064](https://jiraeu.epam.com/browse/EPMDEDP-11064)

### Code Refactoring

- Use common library as dependencies [EPMDEDP-11008](https://jiraeu.epam.com/browse/EPMDEDP-11008)
- Use codebasebranch name from interceptor [EPMDEDP-11031](https://jiraeu.epam.com/browse/EPMDEDP-11031)
- Cobebasebranch has format codebase-gitbranch [EPMDEDP-11057](https://jiraeu.epam.com/browse/EPMDEDP-11057)


<a name="v0.1.5"></a>
## [v0.1.5] - 2022-11-22
### Features

- Allow PIP to search through private repo [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)

### Bug Fixes

- Modify Python PIP auth [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)

### Code Refactoring

- Introduce common library [EPMDEDP-11008](https://jiraeu.epam.com/browse/EPMDEDP-11008)
- Update gerrit-ssh-cmd task [EPMDEDP-11031](https://jiraeu.epam.com/browse/EPMDEDP-11031)


<a name="v0.1.4"></a>
## [v0.1.4] - 2022-11-18
### Features

- Implement mkdocs task [EPMDEDP-10877](https://jiraeu.epam.com/browse/EPMDEDP-10877)
- Add custom helm-push-gh-pages Task [EPMDEDP-10878](https://jiraeu.epam.com/browse/EPMDEDP-10878)
- Set CommitMessagePattern empty string if nil [EPMDEDP-11023](https://jiraeu.epam.com/browse/EPMDEDP-11023)
- Populate PipelineRun with EDP labels [EPMDEDP-11031](https://jiraeu.epam.com/browse/EPMDEDP-11031)

### Bug Fixes

- Search codebase by gitUrlPath with slash [EPMDEDP-10969](https://jiraeu.epam.com/browse/EPMDEDP-10969)

### Routine

- Add commitMessagePattern to TriggerBinding [EPMDEDP-10647](https://jiraeu.epam.com/browse/EPMDEDP-10647)


<a name="v0.1.3"></a>
## [v0.1.3] - 2022-11-16
### Features

- Add crdocs task [EPMDEDP-10872](https://jiraeu.epam.com/browse/EPMDEDP-10872)
- Add Task for updating the version and appVersion of  Helm Chart [EPMDEDP-10879](https://jiraeu.epam.com/browse/EPMDEDP-10879)

### Bug Fixes

- Fix label variable for custom chart [EPMDEDP-10563](https://jiraeu.epam.com/browse/EPMDEDP-10563)
- Remove commitMessagePattern from TriggerBinding [EPMDEDP-10647](https://jiraeu.epam.com/browse/EPMDEDP-10647)

### Routine

- Add Gitlab Maven Java Apps and Libs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Remove parameters from init-values [EPMDEDP-10647](https://jiraeu.epam.com/browse/EPMDEDP-10647)
- Change VERSION to IS_TAG for IMAGE env of kaniko-build Task [EPMDEDP-10839](https://jiraeu.epam.com/browse/EPMDEDP-10839)
- Change VERSION to IS_TAG for IMAGE_TAG env of update-cbis Task [EPMDEDP-10839](https://jiraeu.epam.com/browse/EPMDEDP-10839)
- Set stable version for edp-tekton dependency [EPMDEDP-11013](https://jiraeu.epam.com/browse/EPMDEDP-11013)


<a name="v0.1.2"></a>
## [v0.1.2] - 2022-11-11
### Features

- Move custom logic out of Core pipelines [EPMDEDP-10563](https://jiraeu.epam.com/browse/EPMDEDP-10563)
- Add initial structure for custom pipelines [EPMDEDP-10563](https://jiraeu.epam.com/browse/EPMDEDP-10563)
- Add .Net libraries [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
- Add Java8 and Java libs to Tekton [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
- Add JavaScript libs to Tekton [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
- Add Python libraries [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
- Add terraform libs [EPMDEDP-10595](https://jiraeu.epam.com/browse/EPMDEDP-10595)
- Implement opa libs [EPMDEDP-10597](https://jiraeu.epam.com/browse/EPMDEDP-10597)
- Add custom npm-push task [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Add custom Tekton dotnet agent [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Add tekton EDPComponent [EPMDEDP-10801](https://jiraeu.epam.com/browse/EPMDEDP-10801)
- Add Tekton headlamp pipeline draft [EPMDEDP-10870](https://jiraeu.epam.com/browse/EPMDEDP-10870)
- Add Tekton operator pipelines draft [EPMDEDP-10871](https://jiraeu.epam.com/browse/EPMDEDP-10871)
- Use gitUrlPath to return codebase name [EPMDEDP-10969](https://jiraeu.epam.com/browse/EPMDEDP-10969)

### Bug Fixes

- Interceptor security issues [EPMDEDP-10735](https://jiraeu.epam.com/browse/EPMDEDP-10735)
- Parse request payload from GitHub [EPMDEDP-10837](https://jiraeu.epam.com/browse/EPMDEDP-10837)
- Modify go task to pass images from pipelines [EPMDEDP-10871](https://jiraeu.epam.com/browse/EPMDEDP-10871)
- Interceptor is failed with panic when the framework is not defined in the codebase spec [EPMDEDP-10984](https://jiraeu.epam.com/browse/EPMDEDP-10984)

### Code Refactoring

- Change Docker images [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Remove webhook provisioning [EPMDEDP-10743](https://jiraeu.epam.com/browse/EPMDEDP-10743)
- Create tekton service account [EPMDEDP-10796](https://jiraeu.epam.com/browse/EPMDEDP-10796)
- Rename git-tag to git-cli task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Modify golang task for custom pipelines [EPMDEDP-10871](https://jiraeu.epam.com/browse/EPMDEDP-10871)
- Modify helm-lint task for custom pipelines [EPMDEDP-10876](https://jiraeu.epam.com/browse/EPMDEDP-10876)
- Merge all get-nexus-repository-url Tasks into one [EPMDEDP-11003](https://jiraeu.epam.com/browse/EPMDEDP-11003)

### Routine

- Update sonar gradle plugin [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
- Add support for Build and Review pipelines of Autotests [EPMDEDP-10598](https://jiraeu.epam.com/browse/EPMDEDP-10598)
- Add library build trigger into the GitHub EventLisener [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Align Tekton Triggers of GitHub and Maven pipelines [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Gradle for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of React for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Python for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Go for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Dotnet for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Docker for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Maven and Gradle app and lib with Java8 for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of OPA for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Groovy for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Python lib for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of React lib for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Dotnet lib for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Autotests for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add autotests build trigger into the GitHub EventLisener [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Autotest with Java8 for GitLab vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Maven and Gradle libs with Java11 for GitLab vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Build and Review pipelines of Terraform lib for GitHub vcs [EPMDEDP-10605](https://jiraeu.epam.com/browse/EPMDEDP-10605)
- Add Gitlab Gradle Java Apps and Libs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Python Apps and Libs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Dotnet  Apps and Libs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Docker Lib [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Javascript Apps and Libs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Run pytests in parallel by pytest-xdist [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab OPA Lib [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Terraform Lib [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Groovy Lib [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Gitlab Autotests Java [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add Build and Review pipelines of Go app for GitLab vcs [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Add autotests, lib  build trigger into the GitLab EventListener [EPMDEDP-10606](https://jiraeu.epam.com/browse/EPMDEDP-10606)
- Change openssh image link [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Update npm images [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Change Docker images [EPMDEDP-10664](https://jiraeu.epam.com/browse/EPMDEDP-10664)
- Added support for using branches with a slash (Buildtool Groovy) [EPMDEDP-10736](https://jiraeu.epam.com/browse/EPMDEDP-10736)
- Added support for using branches with a slash (Buildtool Dotnet) [EPMDEDP-10736](https://jiraeu.epam.com/browse/EPMDEDP-10736)
- Add support for using branches with a slash (Buildtools Maven, ReactJs, Python, Go) [EPMDEDP-10736](https://jiraeu.epam.com/browse/EPMDEDP-10736)
- Added support for using branches with a slash (Buildtool Gradle) [EPMDEDP-10736](https://jiraeu.epam.com/browse/EPMDEDP-10736)
- Set GOCACHE parameter as workspace path [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Change localRepository parameter and add HOME env [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Split the code-review trigger for Gerrit and GitHub [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME env and use it as safe.directory for git [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Set '/tekton/home' to USER_HOME parameter of git-cli Task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Use custom tfenv docker image for terraform [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME env into Python Task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME and SEMGREP_VERSION_CACHE_PATH envs to SAST task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME env into Dotnet Task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME env into Codenarc Task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add HOME parameter into npm Task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add parameters into Gradle pipeline and task [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add AWS_DEFAULT_REGION parameter [EPMDEDP-10803](https://jiraeu.epam.com/browse/EPMDEDP-10803)
- Add custom get-version Task for docker, helm, platform [EPMDEDP-10875](https://jiraeu.epam.com/browse/EPMDEDP-10875)
- Update dependency template [EPMDEDP-10969](https://jiraeu.epam.com/browse/EPMDEDP-10969)


<a name="v0.1.1"></a>
## [v0.1.1] - 2022-10-06
### Features

- Add Gerrit CodeReview/Build EventListeners [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Update CodeReview Pipeline checkout for Gerrit VCS [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Implement gerrit voting from pipelines [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Put link to Tekton PipelineRun in Gerrit [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Put gerrit message for build pipeline [EPMDEDP-10428](https://jiraeu.epam.com/browse/EPMDEDP-10428)
- Add GitHub integration [EPMDEDP-10429](https://jiraeu.epam.com/browse/EPMDEDP-10429)
- Create webhook in GitLab [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Add support for GitLab [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Add EDP getversion step [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Update Tasks and Pipelines [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Add Build Pipeline with Gradle build tool [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Use nexus from EDP as default python pypi mirror [EPMDEDP-10432](https://jiraeu.epam.com/browse/EPMDEDP-10432)
- Implement Review/Build pipeline for Python Application [EPMDEDP-10432](https://jiraeu.epam.com/browse/EPMDEDP-10432)
- Implement GO Build and Code Review Tekton pipelines [EPMDEDP-10433](https://jiraeu.epam.com/browse/EPMDEDP-10433)
- Enable gerrit support for JavaScript Application [EPMDEDP-10434](https://jiraeu.epam.com/browse/EPMDEDP-10434)
- Add java11 build pipeline [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Add default and edp versioning type into the pipelines [EPMDEDP-10506](https://jiraeu.epam.com/browse/EPMDEDP-10506)
- Add pruner for pipelineruns and taskruns [EPMDEDP-10509](https://jiraeu.epam.com/browse/EPMDEDP-10509)
- Add Sonar scanning for code-review pipelines [EPMDEDP-10535](https://jiraeu.epam.com/browse/EPMDEDP-10535)
- Implement EDP interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Add deployment templates for EDP Interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Generalize Review/Build Pipelines for Java11 [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Implement github/gitlab events processing [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Add commit-validate task [EPMDEDP-10588](https://jiraeu.epam.com/browse/EPMDEDP-10588)
- Implement Tekton dockerbuild-verify task [EPMDEDP-10589](https://jiraeu.epam.com/browse/EPMDEDP-10589)
- Implement helm-docs task [EPMDEDP-10592](https://jiraeu.epam.com/browse/EPMDEDP-10592)
- Build and Code-review pipelines for Dotnet [EPMDEDP-10593](https://jiraeu.epam.com/browse/EPMDEDP-10593)
- Add ecr-to-docker task [EPMDEDP-10594](https://jiraeu.epam.com/browse/EPMDEDP-10594)
- Add push-to-jira Task [EPMDEDP-10596](https://jiraeu.epam.com/browse/EPMDEDP-10596)
- Add unit tests for interceptor [EPMDEDP-10599](https://jiraeu.epam.com/browse/EPMDEDP-10599)
- Implement HTTPS connection [EPMDEDP-10600](https://jiraeu.epam.com/browse/EPMDEDP-10600)
- Implement Review/Build pipelines for Containers type Library [EPMDEDP-10603](https://jiraeu.epam.com/browse/EPMDEDP-10603)
- Implement Groovy library build and review [EPMDEDP-10604](https://jiraeu.epam.com/browse/EPMDEDP-10604)
- Separate applications and libraries [EPMDEDP-10604](https://jiraeu.epam.com/browse/EPMDEDP-10604)
- Add get-nuget-token Task for Dotnet applications [EPMDEDP-10652](https://jiraeu.epam.com/browse/EPMDEDP-10652)

### Bug Fixes

- Updated tasks dependency between each other [EPMDEDP-10535](https://jiraeu.epam.com/browse/EPMDEDP-10535)
- Align Build and Review pipeline name by common pattern [EPMDEDP-10535](https://jiraeu.epam.com/browse/EPMDEDP-10535)
- Remove service account for task in gerrit review pipeline [EPMDEDP-10535](https://jiraeu.epam.com/browse/EPMDEDP-10535)
- Fix deployment template mapping [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Fix EDP interceptor deployment configuration [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Add CODEBASE_NAME parameter to dotnet review [EPMDEDP-10589](https://jiraeu.epam.com/browse/EPMDEDP-10589)
- Add build stage to gradle review [EPMDEDP-10589](https://jiraeu.epam.com/browse/EPMDEDP-10589)
- Remove service account for tasks in review pipeline [EPMDEDP-10589](https://jiraeu.epam.com/browse/EPMDEDP-10589)
- Change Kaniko repository name pattern [EPMDEDP-10594](https://jiraeu.epam.com/browse/EPMDEDP-10594)
- Fix port for interceptor service [EPMDEDP-10600](https://jiraeu.epam.com/browse/EPMDEDP-10600)
- Log correct build info [EPMDEDP-10600](https://jiraeu.epam.com/browse/EPMDEDP-10600)
- Remove Kaniko enabled parameter [EPMDEDP-10604](https://jiraeu.epam.com/browse/EPMDEDP-10604)
- Various fixes and refactor Sonar tasks [EPMDEDP-10604](https://jiraeu.epam.com/browse/EPMDEDP-10604)

### Code Refactoring

- Refactor GitHub WebHook creation [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Move github-set-status task to tasks dir [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Rename files for tekton-resources [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Move secrets under resource section [EPMDEDP-10430](https://jiraeu.epam.com/browse/EPMDEDP-10430)
- Service account edp-kinako might be already created [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Create PVC during pipelineRun [EPMDEDP-10431](https://jiraeu.epam.com/browse/EPMDEDP-10431)
- Typo in abbreviation VCS [EPMDEDP-10434](https://jiraeu.epam.com/browse/EPMDEDP-10434)
- Switch to nexus in specific namespace [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Update labels [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Introduce helm-chart as installer [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Remove specific namespace definition from resource [EPMDEDP-10435](https://jiraeu.epam.com/browse/EPMDEDP-10435)
- Use generalized approach for go tool [EPMDEDP-10586](https://jiraeu.epam.com/browse/EPMDEDP-10586)
- Make sonar-cleanup common [EPMDEDP-10590](https://jiraeu.epam.com/browse/EPMDEDP-10590)
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
- Add tests for pruner section in charts [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add unittests for helm templates [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add unittests for charts baseline [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add unittests for pipelines [EPMDEDP-10570](https://jiraeu.epam.com/browse/EPMDEDP-10570)
- Add Dockerfile for edp-interceptor [EPMDEDP-10582](https://jiraeu.epam.com/browse/EPMDEDP-10582)
- Add github actions [EPMDEDP-10603](https://jiraeu.epam.com/browse/EPMDEDP-10603)
- Disable code coverage for main.go [EPMDEDP-10650](https://jiraeu.epam.com/browse/EPMDEDP-10650)

### Documentation

- Update README file with general information [EPMDEDP-10649](https://jiraeu.epam.com/browse/EPMDEDP-10649)


<a name="v0.1.0"></a>
## v0.1.0 - 2022-08-23

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.5.0...HEAD
[v0.5.0]: https://github.com/epam/edp-tekton/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/epam/edp-tekton/compare/v0.3.1...v0.4.0
[v0.3.1]: https://github.com/epam/edp-tekton/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/epam/edp-tekton/compare/v0.2.9...v0.3.0
[v0.2.9]: https://github.com/epam/edp-tekton/compare/v0.2.8...v0.2.9
[v0.2.8]: https://github.com/epam/edp-tekton/compare/v0.2.7...v0.2.8
[v0.2.7]: https://github.com/epam/edp-tekton/compare/v0.2.6...v0.2.7
[v0.2.6]: https://github.com/epam/edp-tekton/compare/v0.2.5...v0.2.6
[v0.2.5]: https://github.com/epam/edp-tekton/compare/v0.2.4...v0.2.5
[v0.2.4]: https://github.com/epam/edp-tekton/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/epam/edp-tekton/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/epam/edp-tekton/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/epam/edp-tekton/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/epam/edp-tekton/compare/v0.1.9...v0.2.0
[v0.1.9]: https://github.com/epam/edp-tekton/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/epam/edp-tekton/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/epam/edp-tekton/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/epam/edp-tekton/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/epam/edp-tekton/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/epam/edp-tekton/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/epam/edp-tekton/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/epam/edp-tekton/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/epam/edp-tekton/compare/v0.1.0...v0.1.1
