Overview
========
dockergen is a tool that programmatically builds and publishes Docker images based on Docker templates and variable
substitution using declarative configuration. This tool is useful for cases in which there are multiple Docker templates
that should be executed with one variable value varied. Examples of this include building a single template for multiple
versions of Java or Go.

Installation
============
dockergen can be installed using `go get`:

```bash
go get github.com/nmiyake/dockergen
```

Examples
========
Consider the case of building a Java Docker image based on both Java 7 and Java 8.

This is the Dockerfile for building a Docker image with an unlimited JCE that also has some utilities for Java 8:

```
FROM davidcaste/alpine-java-unlimited-jce:jdk8

RUN apk add --no-cache \
    bash \
    libstdc++ \
    git \
    openssh \
    tar \
    gzip
```

The Dockerfile for building a Java 7 version of the above is the exact same, except that the final tag is "jdk7" instead
of "jdk8". Although it would be fairly simple to simply duplicate the Dockerfile, this can be a pain if there are
multiple supporting files or if there are many versions for which this should be built.

dockergen can be used to solve this as follows:

* Create `Dockerfile_template.txt` with the following content:

```
FROM davidcaste/alpine-java-unlimited-jce:{{.jdkVersion}}

RUN apk add --no-cache \
    bash \
    libstdc++ \
    git \
    openssh \
    tar \
    gzip
```

* Create a `config.yml` with the following content:

```
build-id-var: CIRCLE_BUILD_NUM
tag-suffix: -t{{BuildID}}
for:
  jdkVersion:
    - jdk7
    - jdk8
builds:
  unlimited-jce:
    docker-template: Dockerfile_template.txt
    tag: nmiyake/alpine-java-unlimited-jce:{{.jdkVersion}}
```

The `build-id-var` specifies the name of the environment variable that contains a unique identifier that can be used to
identify builds. When a build is run in CircleCI, the `CIRCLE_BUILD_NUM` environment variable contains the build number,
so this configuration uses that as a unique build identifier.

Assuming that `CIRCLE_BUILD_NUM` has a value of 13, running `dockergen --config config.yml build` will build and tag
`nmiyake/alpine-java-unlimited-jce:jdk7-t13` and `nmiyake/alpine-java-unlimited-jce:jdk8-t13` with the provided
template, where the Dockerfiles that are built are the template with the template value replaced with the value for that
execution.

It is possible to define multiple builds in the `builds` block. It is also possible to define a `for` block within a
`build` block (in which case the `for` loop will only execute for that build). It is also possible to define multiple
variables that are cycled over in a `for` block (each named variable must have the same number of elements).

License
=======
This project is made available under the [MIT License](https://opensource.org/licenses/MIT).
