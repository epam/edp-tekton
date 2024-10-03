<a name="unreleased"></a>
## [Unreleased]

### Features

- Add Support for BitBucket as GitServer ([#317](https://github.com/epam/edp-tekton/issues/317))
- Add bitbucket-set-status task for tekton pipelines ([#315](https://github.com/epam/edp-tekton/issues/315))
- Add Tekton config to support for BitBucket Cloud ([#311](https://github.com/epam/edp-tekton/issues/311))
- Add support for BitBucket Cloud ([#311](https://github.com/epam/edp-tekton/issues/311))
- Add maven-rpm pipelines ([#282](https://github.com/epam/edp-tekton/issues/282))
- Add deploy pipeline with approve task([#279](https://github.com/epam/edp-tekton/issues/279))
- Implement rpm build functionality ([#272](https://github.com/epam/edp-tekton/issues/272))
- Enhance Pipeline Triggering Process ([#263](https://github.com/epam/edp-tekton/issues/263))
- Add dynamic pipeline name retrieval from CodebaseBranch spec ([#263](https://github.com/epam/edp-tekton/issues/263))
- Add C make/cmake application pipelines ([#260](https://github.com/epam/edp-tekton/issues/260))
- Update jira task Include GitCommitUrl instead of Pipelinerun([#252](https://github.com/epam/edp-tekton/issues/252))
- Add clean pipelines for tekton ([#244](https://github.com/epam/edp-tekton/issues/244))
- Re-trigger Tekton Pipeline with /ok-to-test comment ([#238](https://github.com/epam/edp-tekton/issues/238))
- Add GitHub owners check configuration ([#234](https://github.com/epam/edp-tekton/issues/234))
- Add pipelines tasks tests for new codebase ansible([#236](https://github.com/epam/edp-tekton/issues/236))

### Bug Fixes

- Fix c lang pipelines ([#290](https://github.com/epam/edp-tekton/issues/290))
- Fix TriggerTemplate for deploy with approve step ([#279](https://github.com/epam/edp-tekton/issues/279))
- Align rpm-build argo diff  ([#272](https://github.com/epam/edp-tekton/issues/272))
- Update parameter definition ([#261](https://github.com/epam/edp-tekton/issues/261))
- Fix code duplication in dockerbuild-verify Task([#242](https://github.com/epam/edp-tekton/issues/242))
- Fix commit-validate task to check only the first line ([#229](https://github.com/epam/edp-tekton/issues/229))
- Fix commit-validate task to check only the first line ([#229](https://github.com/epam/edp-tekton/issues/229))
- Install packages via requirements ([#215](https://github.com/epam/edp-tekton/issues/215))

### Routine

- Update e2e tests for GH, switch to stable version of Tekton ([#311](https://github.com/epam/edp-tekton/issues/311))
- Remove kaniko cache usage ([#309](https://github.com/epam/edp-tekton/issues/309))
- Align tasks field in test ([#290](https://github.com/epam/edp-tekton/issues/290))
- Validate 'type' field for QuickLink resources ([#304](https://github.com/epam/edp-tekton/issues/304))
- Format Tekton Tasks and Pipelines According to Pre-defined Structure ([#290](https://github.com/epam/edp-tekton/issues/290))
- Remove config.yaml from validation script ([#290](https://github.com/epam/edp-tekton/issues/290))
- Add description to the approval promote procedure ([#279](https://github.com/epam/edp-tekton/issues/279))
- Update save-cache and get-cache tasks ([#294](https://github.com/epam/edp-tekton/issues/294))
- Update base image for cache tasks ([#294](https://github.com/epam/edp-tekton/issues/294))
- Switch cache compression algorithm to zstd ([#294](https://github.com/epam/edp-tekton/issues/294))
- Add check for the Tekton Pipeline and Task structure ([#290](https://github.com/epam/edp-tekton/issues/290))
- Align Tekton Piplines formating
- Add k8s 1.30 check and update kuttle ([#282](https://github.com/epam/edp-tekton/issues/282))
- Update Tekton-dashboard version to v0.50.0 ([#286](https://github.com/epam/edp-tekton/issues/286))
- Do not set make parameters for rpm build ([#282](https://github.com/epam/edp-tekton/issues/282))
- Update rpm-build flow for java17 maven ([#282](https://github.com/epam/edp-tekton/issues/282))
- Update report message
- Enable labels for review and build pipelines ([#270](https://github.com/epam/edp-tekton/issues/270))
- Disable Tekton dashboard deployments([#266](https://github.com/epam/edp-tekton/issues/266))
- Update KubeRocketCI names ([#258](https://github.com/epam/edp-tekton/issues/258))
- Update templates to include changeNumber from Merge Request ([#250](https://github.com/epam/edp-tekton/issues/250))
- Make Tekton Dashboard quickLink deployment optional([#246](https://github.com/epam/edp-tekton/issues/246))
- Update Gerrit Trigger Template for Remote Cluster Deployment ([#248](https://github.com/epam/edp-tekton/issues/248))
- Update github.com/epam/edp-codebase-operator dependency ([#240](https://github.com/epam/edp-tekton/issues/240))
- Switch redirect link from Tekton Dashboard to KRCI Portal page ([#232](https://github.com/epam/edp-tekton/issues/232))
- Add Grafana Tekton dashboard ([#227](https://github.com/epam/edp-tekton/issues/227))
- Implement Results Emission for All Build Pipelines ([#225](https://github.com/epam/edp-tekton/issues/225))
- Revert Update gitservers event listener name and add label ([#222](https://github.com/epam/edp-tekton/issues/222))
- Update gitservers event listener name and add label ([#222](https://github.com/epam/edp-tekton/issues/222))
- Align to new name KubeRocketCI ([#220](https://github.com/epam/edp-tekton/issues/220))
- Address flaky e2e tests ([#215](https://github.com/epam/edp-tekton/issues/215))
- Update tekton-helm image version ([#215](https://github.com/epam/edp-tekton/issues/215))
- Update current version ([#211](https://github.com/epam/edp-tekton/issues/211))
- Update kuttle and kind images ([#215](https://github.com/epam/edp-tekton/issues/215))

### Documentation

- Update CHANGELOG md ([#302](https://github.com/epam/edp-tekton/issues/302))
- Replace EDP with KubeRocketCI in README md ([#268](https://github.com/epam/edp-tekton/issues/268))


<a name="v0.12.0"></a>
## [v0.12.0] - 2024-06-13
### Features

- Align chart for ingress TLS configuration([#178](https://github.com/epam/edp-tekton/issues/178))
- Add lastCommitMessage to interceptor response ([#193](https://github.com/epam/edp-tekton/issues/193))
- Add quality gate for chart name alignment with codebase name([#191](https://github.com/epam/edp-tekton/issues/191))
- Dynamically set repository URLs for package types ([#132](https://github.com/epam/edp-tekton/issues/132))
- Enhance Gradle proxy support ([#132](https://github.com/epam/edp-tekton/issues/132))
- Exclude pipelinerun from resources displayed in argocd([#169](https://github.com/epam/edp-tekton/issues/169))

### Bug Fixes

- Resolve Version Conflict Between h11 and httpcore ([#195](https://github.com/epam/edp-tekton/issues/195))
- Set full stage name in autotests tekton task ([#199](https://github.com/epam/edp-tekton/issues/199))
- Make possible work with registry without registry parameter([#184](https://github.com/epam/edp-tekton/issues/184))
- multiple eventListeners route creation([#175](https://github.com/epam/edp-tekton/issues/175))
- Add lint config and remove cache from e2e ([#164](https://github.com/epam/edp-tekton/issues/164))
- Remove volume workspace from commit-validate task ([#78](https://github.com/epam/edp-tekton/issues/78))
- Update argocd-integration step logic ([#153](https://github.com/epam/edp-tekton/issues/153))
- Update custom deploy pipeline logic ([#153](https://github.com/epam/edp-tekton/issues/153))
- Update route for el ([#151](https://github.com/epam/edp-tekton/issues/151))
- Fix gitProviders parameter in custom pipelines ([#150](https://github.com/epam/edp-tekton/issues/150))

### Code Refactoring

- Refactor pipeline and remove nexus URL step ([#132](https://github.com/epam/edp-tekton/issues/132)) ([#177](https://github.com/epam/edp-tekton/issues/177))

### Routine

- Update current version ([#211](https://github.com/epam/edp-tekton/issues/211))
- Update current version ([#211](https://github.com/epam/edp-tekton/issues/211))
- Align argo diff for pruner ([#209](https://github.com/epam/edp-tekton/issues/209))
- Align argo diff for pruner ([#205](https://github.com/epam/edp-tekton/issues/205))
- Update tekton pruner logic ([#205](https://github.com/epam/edp-tekton/issues/205))
- Set default branch during project creation ([#207](https://github.com/epam/edp-tekton/issues/207))
- Update CodeQL to the latest version ([#203](https://github.com/epam/edp-tekton/issues/203))
- Bump Node image version to 18.20.3-alpine3.20 ([#201](https://github.com/epam/edp-tekton/issues/201))
- Use commit message instead PR title ([#197](https://github.com/epam/edp-tekton/issues/197))
- Remove PR modify action for review pipeline ([#187](https://github.com/epam/edp-tekton/issues/187))
- Use Go 1.22 for e2e workflow ([#182](https://github.com/epam/edp-tekton/issues/182))
- Bump hadolint version to v2.12.0-alpine ([#188](https://github.com/epam/edp-tekton/issues/188))
- Bump to go 1.22 ([#182](https://github.com/epam/edp-tekton/issues/182))
- Update tekton-dashboard([#180](https://github.com/epam/edp-tekton/issues/180))
- Switch cache to recreate strategy ([#171](https://github.com/epam/edp-tekton/issues/171))
- Adjust tekton pruner logic ([#147](https://github.com/epam/edp-tekton/issues/147))
- Add codeowners file to the repo ([#166](https://github.com/epam/edp-tekton/issues/166))
- Align commit message pattern ([#160](https://github.com/epam/edp-tekton/issues/160))
- Update sonar project properties ([#160](https://github.com/epam/edp-tekton/issues/160))
- Remove mount volume from all commit-validate tasks([#159](https://github.com/epam/edp-tekton/issues/159))
- Update workspace volume size ([#158](https://github.com/epam/edp-tekton/issues/158))
- Align codebase branch name and codebase name([#157](https://github.com/epam/edp-tekton/issues/157))
- Update Tekton pruner logic ([#147](https://github.com/epam/edp-tekton/issues/147))
- Bump custom deploy images ([#153](https://github.com/epam/edp-tekton/issues/153))
- Add timeout after clean-edp task ([#152](https://github.com/epam/edp-tekton/issues/152))
- Switch Argo CD integration to edp-ci user ([#152](https://github.com/epam/edp-tekton/issues/152))
- Use new icon for gerrit QuickLink ([#152](https://github.com/epam/edp-tekton/issues/152))
- Remove deprecated EDPComponents CRD ([#149](https://github.com/epam/edp-tekton/issues/149))
- Bump java-maven deploy plugin version ([#148](https://github.com/epam/edp-tekton/issues/148))
- Update current development version ([#149](https://github.com/epam/edp-tekton/issues/149))


<a name="v0.11.0"></a>
## [v0.11.0] - 2024-03-12
### Features

- Make possible run autotest from deploy pipeline ([#133](https://github.com/epam/edp-tekton/issues/133))
- Add Github CI workflow for PR ([#136](https://github.com/epam/edp-tekton/issues/136))
- Make possible run post/pre deploy in remote cluster ([#133](https://github.com/epam/edp-tekton/issues/133))
- Enable custom values parameter ([#131](https://github.com/epam/edp-tekton/issues/131))
- Set GitHub as default gitProvider ([#130](https://github.com/epam/edp-tekton/issues/130))
- Add support for multiple GitProviders ([#130](https://github.com/epam/edp-tekton/issues/130))
- Integrate DotNet Pipelines with custom registry ([#127](https://github.com/epam/edp-tekton/issues/127))
- Add nodeSelector, affinity, tolerations ([#126](https://github.com/epam/edp-tekton/issues/126))
- Enable correct deployment name for tekton cache chart ([#126](https://github.com/epam/edp-tekton/issues/126))
- Integrate Python Pipelines with custom registry ([#123](https://github.com/epam/edp-tekton/issues/123))
- Implement custom pipeline for clean edp ([#117](https://github.com/epam/edp-tekton/issues/117))
- Add support for argocd app wait deployment ([#117](https://github.com/epam/edp-tekton/issues/117))
- Add QuickLink Custom Resources ([#114](https://github.com/epam/edp-tekton/issues/114))
- Integrate NPM Pipelines with custom registry ([#115](https://github.com/epam/edp-tekton/issues/115))
- Integrate Antora Pipelines with custom registry ([#115](https://github.com/epam/edp-tekton/issues/115))
- Add integration tests step for nexus-operator ([#116](https://github.com/epam/edp-tekton/issues/116))
- Add ability to use custom Maven settings.xml ([#106](https://github.com/epam/edp-tekton/issues/106))
- Migrate custom Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate Java-Gradle Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate Java-Maven Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate Dotnet Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate Go Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate NPM Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Migrate Python Tekton pipelines for new Sonar branch approach([#107](https://github.com/epam/edp-tekton/issues/107))
- Add secret for authentication in package registries ([#106](https://github.com/epam/edp-tekton/issues/106))
- Add pull request data to interceptor response ([#105](https://github.com/epam/edp-tekton/issues/105))
- Provision default TriggerTemplate for CD Pipeline ([#96](https://github.com/epam/edp-tekton/issues/96))
- Align promotion procedure to the new format ([#96](https://github.com/epam/edp-tekton/issues/96))
- Implement deploy pipeline based on Argo ApplicationSet ([#96](https://github.com/epam/edp-tekton/issues/96))
- Add custom pipelines for SAM tool ([#92](https://github.com/epam/edp-tekton/issues/92))

### Bug Fixes

- Update customValues flag operation logic ([#133](https://github.com/epam/edp-tekton/issues/133))
- Update autotest-runner ([#146](https://github.com/epam/edp-tekton/issues/146))
- Invalid CodebaseImageStream tag date format ([#143](https://github.com/epam/edp-tekton/issues/143))
- Fix git server hostname extraction for GitLab ([#144](https://github.com/epam/edp-tekton/issues/144))
- Update resource creation logic ([#142](https://github.com/epam/edp-tekton/issues/142))
- Align release pipeline kaniko task([#140](https://github.com/epam/edp-tekton/issues/140))
- We must pass context with deployable module for kaniko ([#141](https://github.com/epam/edp-tekton/issues/141))
- Run sonar after integration-test for keycloak-operator ([#139](https://github.com/epam/edp-tekton/issues/139))
- Update task dependencies ([#136](https://github.com/epam/edp-tekton/issues/136))
- Enable uniq selector pattern for deploy pipeline ([#137](https://github.com/epam/edp-tekton/issues/137))
- Add https prefix into webhookUrl ([#130](https://github.com/epam/edp-tekton/issues/130))
- Align promote images to the new payload pattern([#134](https://github.com/epam/edp-tekton/issues/134))
- Add kubeconfig parameter to deploy TT ([#133](https://github.com/epam/edp-tekton/issues/133))
- Ensure build pipeline is triggered only for Merge Events ([#128](https://github.com/epam/edp-tekton/issues/128))
- Update working directory path for custom pipelines ([#119](https://github.com/epam/edp-tekton/issues/119))
- Fix ingress annotation in Tekton dashboard ([#111](https://github.com/epam/edp-tekton/issues/111))
- Align parameters name for CD Pipeline ([#96](https://github.com/epam/edp-tekton/issues/96))
- Align parameters name for CD Pipeline ([#96](https://github.com/epam/edp-tekton/issues/96))
- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))
- Change versioning for helm lib default([#101](https://github.com/epam/edp-tekton/issues/101))
- Add chart-dir parameter to helm pipeline([#101](https://github.com/epam/edp-tekton/issues/101))
- Add extraline in private ssh key ([#100](https://github.com/epam/edp-tekton/issues/100))
- Failed push-to-jira step in build pipeline on okd ([#94](https://github.com/epam/edp-tekton/issues/94))
- Fix Service name for Ingress object of the EventListeners CR ([#93](https://github.com/epam/edp-tekton/issues/93))
- Align cache endpoint with service name ([#89](https://github.com/epam/edp-tekton/issues/89))
- Fix tekton cache service name ([#89](https://github.com/epam/edp-tekton/issues/89))

### Code Refactoring

- Remove deprecated autotests functional ([#145](https://github.com/epam/edp-tekton/issues/145))
- Align TriggerTemplate stage label name ([#145](https://github.com/epam/edp-tekton/issues/145))
- Align labels, name convention ([#145](https://github.com/epam/edp-tekton/issues/145))
- Switch to using gitServer name instead of gitProvider ([#130](https://github.com/epam/edp-tekton/issues/130))
- Change gitServer structure ([#130](https://github.com/epam/edp-tekton/issues/130))
- Refactor Tekton Triggers components deployment ([#130](https://github.com/epam/edp-tekton/issues/130))
- Simplify Git provider EventListeners and Ingress handling ([#93](https://github.com/epam/edp-tekton/issues/93))

### Testing

- Run e2e tests one-by-one ([#130](https://github.com/epam/edp-tekton/issues/130))

### Routine

- Update current version ([#149](https://github.com/epam/edp-tekton/issues/149))
- Bump helm-docs image version ([#149](https://github.com/epam/edp-tekton/issues/149))
- Bump alpine version ([#149](https://github.com/epam/edp-tekton/issues/149))
- Allow to define webhook URL for Github/Gitlab ([#130](https://github.com/epam/edp-tekton/issues/130))
- Add parameters for gerrit gitserver ([#136](https://github.com/epam/edp-tekton/issues/136))
- Implement cache in antora pipeline ([#138](https://github.com/epam/edp-tekton/issues/138))
- Update github workflow ([#136](https://github.com/epam/edp-tekton/issues/136))
- Get webhook url of gitlab and github from gitserver resource ([#130](https://github.com/epam/edp-tekton/issues/130))
- Add label to associate an ingress with a specific GitServer ([#130](https://github.com/epam/edp-tekton/issues/130))
- Allow overriding fields in gitServer and QuickLink CRs ([#168](https://github.com/epam/edp-tekton/issues/168))
- Remove Dashboard EDPComponent CR ([#168](https://github.com/epam/edp-tekton/issues/168))
- Remove unused tasks from DotNet Pipelines ([#127](https://github.com/epam/edp-tekton/issues/127))
- Remove unused get-nexus-repository-url Task from Python Pipelines ([#123](https://github.com/epam/edp-tekton/issues/123))
- Remove 'edp' from helm chart name ([#122](https://github.com/epam/edp-tekton/issues/122))
- Make possible to use cache at python-push step ([#121](https://github.com/epam/edp-tekton/issues/121))
- Increase RAM request and limit of save-cache task ([#120](https://github.com/epam/edp-tekton/issues/120))
- Bump tekton dashboard version ([#118](https://github.com/epam/edp-tekton/issues/118))
- Enable pip package caching ([#150](https://github.com/epam/edp-tekton/issues/150))
- Separating nexus-operator pipelines from general pipelines ([#116](https://github.com/epam/edp-tekton/issues/116))
- Migrate from update-build-number to maven task ([#112](https://github.com/epam/edp-tekton/issues/112))
- Add comments with description for tekton-cache ([#110](https://github.com/epam/edp-tekton/issues/110))
- Enable service name customization in the tekton-cache Helm chart ([#110](https://github.com/epam/edp-tekton/issues/110))
- Bump semgrep image version ([#109](https://github.com/epam/edp-tekton/issues/109))
- Add e2e tests for edp-tekton custom pipeline ([#108](https://github.com/epam/edp-tekton/issues/108))
- Remove unused sonar tasks ([#107](https://github.com/epam/edp-tekton/issues/107))
- Update maven-deploy-plugin syntax for Java 8 and 11 compatibility ([#106](https://github.com/epam/edp-tekton/issues/106))
- Remove unused Nexus-specific properties, GitLab VCS ([#106](https://github.com/epam/edp-tekton/issues/106))
- Remove unused Nexus-specific properties, GitHub VCS ([#106](https://github.com/epam/edp-tekton/issues/106))
- Remove unused Nexus-specific properties, Gerrit VCS ([#106](https://github.com/epam/edp-tekton/issues/106))
- Remove 'get-nexus-repository-url' task ([#106](https://github.com/epam/edp-tekton/issues/106))
- Add examples for using various artifactory storages in Maven ([#106](https://github.com/epam/edp-tekton/issues/106))
- Update current development version ([#104](https://github.com/epam/edp-tekton/issues/104))
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

- Define name convention for ingress objects ([#122](https://github.com/epam/edp-tekton/issues/122))
- Update README md file ([#132](https://github.com/epam/edp-tekton/issues/132))

### BREAKING CHANGE:


We need to change format of payload that we pass
to CD Pipeline


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

[Unreleased]: https://github.com/epam/edp-tekton/compare/v0.12.0...HEAD
[v0.12.0]: https://github.com/epam/edp-tekton/compare/v0.11.0...v0.12.0
[v0.11.0]: https://github.com/epam/edp-tekton/compare/v0.10.2...v0.11.0
[v0.10.2]: https://github.com/epam/edp-tekton/compare/v0.10.1...v0.10.2
[v0.10.1]: https://github.com/epam/edp-tekton/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/epam/edp-tekton/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/epam/edp-tekton/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/epam/edp-tekton/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/epam/edp-tekton/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/epam/edp-tekton/compare/v0.5.0...v0.6.0
