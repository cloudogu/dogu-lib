# Set these to the desired values
PROJECT_NAME=dogu-lib
VERSION=1.0.0
ARTIFACT_ID=dogu_lib

MAKEFILES_VERSION=10.10.0

GOTAG?=1.26.5
GO_BUILD_FLAGS?=-mod=vendor -a ./...
LINT_VERSION=v2.10.1

include build/make/variables.mk
include build/make/self-update.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/release.mk
include build/make/mocks.mk

default: compile
