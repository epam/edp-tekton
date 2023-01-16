# Maven

This Task can be used to run a Maven goals on a simple maven project or on a multi-module maven project.
## Parameters

- **MAVEN_IMAGE**: The base image for maven (_default_: `gcr.io/cloud-builders/mvn`)
- **GOALS**: Maven `goals` to be executed
- **CONTEXT_DIR**: The context directory within the repository for sources on which we want to execute maven goals. (_Default_: ".")
- **ci-secret**: Name of the secret holding the CI maven secret

## Workspaces

- **source**: `PersistentVolumeClaim`-type so that volume can be shared among `git-clone` and `maven` task

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: maven-source-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Mi
```

## Platforms

The Task can be run on `linux/amd64`, `linux/s390x` and `linux/ppc64le` platforms.

For `linux/s390x` and `linux/ppc64le` platforms specify **MAVEN_IMAGE** parameter with `maven:3.6.3-adoptopenjdk-11` value in TaskRun or PipelineRun.

## Usage

This Pipeline and PipelineRun runs a Maven build on a particular module in a multi-module maven project

### With Defaults

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: maven-test-pipeline
spec:
  workspaces:
    - name: shared-workspace
  tasks:
    - name: fetch-repository
      taskRef:
        kind: Task
        name: git-clone
      workspaces:
        - name: output
          workspace: shared-workspace
      params:
        - name: url
          value: https://github.com/redhat-developer-demos/tekton-tutorial
        - name: subdirectory
          value: ""
        - name: deleteExisting
          value: "true"
    - name: maven-run
      taskRef:
        kind: Task
        name: maven
      runAfter:
        - fetch-repository
      params:
        - name: CONTEXT_DIR
          value: "apps/greeter/java/quarkus"
        - name: GOALS
          value:
            - -DskipTests
            - clean
            - package
      workspaces:
        - name: source
          workspace: shared-workspace
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: maven-test-pipeline-run
spec:
  pipelineRef:
    name: maven-test-pipeline
  workspaces:
    - name: shared-workspace
      persistentvolumeclaim:
        claimName: maven-source-pvc
```

---

### With Custom Maven Params

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: maven-test-pipeline
spec:
  workspaces:
    - name: shared-workspace
  tasks:
    - name: fetch-repository
      taskRef:
        kind: Task
        name: git-clone
      workspaces:
        - name: output
          workspace: shared-workspace
      params:
        - name: url
          value: https://github.com/redhat-developer-demos/tekton-tutorial
        - name: subdirectory
          value: ""
        - name: deleteExisting
          value: "true"
    - name: maven-run
      taskRef:
        kind: Task
        name: maven
      runAfter:
        - fetch-repository
      params:
        - name: CONTEXT_DIR
          value: "apps/greeter/java/quarkus"
        - name: GOALS
          value:
            - -DskipTests
            - clean
            - package
      workspaces:
        - name: source
          workspace: shared-workspace
```

`PipelineRun` same as above in case of default values

---

### With Custom /.m2/settings.yaml

A user provided custom `settings.xml` can be used with the Maven Task. To do this we need to mount volume with `settings.xml` on the Maven Task.

Following steps demonstrate the use of a ConfigMap to mount like a volume to Task `settings.xml`.

1. create configmap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-maven-settings
data:
  settings.xml: |
    <?xml version="1.0" encoding="UTF-8"?>
    <settings>
      <mirrors>
        <mirror>
          <id>maven.org</id>
          <name>Default mirror</name>
          <url>http://repo1.maven.org/maven2</url>
          <mirrorOf>central</mirrorOf>
        </mirror>
      </mirrors>
    </settings>
```

or

```bash
oc create configmap custom-maven-settings --from-file=settings.xml
```

2. create `Pipeline` and `PipelineRun`

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: maven-test-pipeline
spec:
  workspaces:
    - name: shared-workspace
  tasks:
    - name: fetch-repository
      taskRef:
        kind: Task
        name: git-clone
      workspaces:
        - name: output
          workspace: shared-workspace
      params:
        - name: url
          value: https://github.com/redhat-developer-demos/tekton-tutorial
        - name: subdirectory
          value: ""
        - name: deleteExisting
          value: "true"
    - name: maven-run
      taskRef:
        kind: Task
        name: maven
      runAfter:
        - fetch-repository
      params:
        - name: CONTEXT_DIR
          value: "apps/greeter/java/quarkus"
        - name: GOALS
          value:
            - -DskipTests
            - clean
            - package
      workspaces:
        - name: source
          workspace: shared-workspace
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: maven-test-pipeline-run
spec:
  pipelineRef:
    name: maven-test-pipeline
  workspaces:
    - name: shared-workspace
      persistentvolumeclaim:
        claimName: maven-source-pvc
```
3. create `Task`

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: maven
  annotations:
    tekton.dev/pipelines.minVersion: "0.12.1"
    tekton.dev/categories: Build Tools
    tekton.dev/tags: build-tool
    tekton.dev/platforms: "linux/amd64,linux/s390x,linux/ppc64le"
spec:
  description: >-
    This Task can be used to run a Maven build.
  workspaces:
    - name: source
      description: The workspace consisting of maven project.
  params:
    - name: MAVEN_IMAGE
      type: string
      description: Maven base image
      default: maven:3.8.6-openjdk-11-slim
    - name: GOALS
      description: maven goals to run
      type: array
      default:
        - "package"
    - name: CONTEXT_DIR
      type: string
      description: >-
        The context directory within the repository for sources on
        which we want to execute maven goals.
      default: "."
    - name: ci-secret
      type: string
      description: name of the secret holding the CI maven secret
      default: nexus-ci.user
    - name: ci-sonar-secret
      type: string
      description: name of the secret holding the CI sonar secret
      default: sonar-ciuser-token
  volumes:
    - name: settings-maven
      configMap:
        name: custom-maven-settings
  steps:
    - name: mvn-goals
      image: $(params.MAVEN_IMAGE)
      volumeMounts:
        - name: settings-maven
          mountPath: /var/configmap
      workingDir: $(workspaces.source.path)/$(params.CONTEXT_DIR)
      command: ["/usr/bin/mvn"]
      args:
        - -s
        - /var/configmap/settings.xml
        - "$(params.GOALS)"
      env:
        - name: HOME
          value: $(workspaces.source.path)
        - name: SONAR_TOKEN
          valueFrom:
            secretKeyRef:
              name: $(params.ci-sonar-secret)
              key: secret
        - name: CI_USERNAME
          valueFrom:
            secretKeyRef:
              name: $(params.ci-secret)
              key: username
        - name: CI_PASSWORD
          valueFrom:
            secretKeyRef:
              name: $(params.ci-secret)
              key: password
```