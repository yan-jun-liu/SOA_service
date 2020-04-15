#######################################################################
# Container image configuration
#REGISTRY=ECR be set
PROJECT_NAME = soa
IMAGE_NAME = upload-service
INIT_TAG = v1.0.0
CONTAINER_PORT=8080
INSTANCE_PORT=8081
REMOTE_TAGS = $(shell git ls-remote --tags origin 2>/dev/null)
TAG = $(shell git describe --tags --abbrev=0 2>/dev/null)

DCO = $(shell git config --global --get user.name) <$(shell git config --global --get user.email)>

define RELEASE_COMMIT_MESSAGE
cd(release): release version {{currentTag}} 

Bump up the release version to {{currentTag}}

Signed-off-by: $(DCO)
endef
export RELEASE_COMMIT_MESSAGE

ifeq ($(TAG),)
TAG := $(INIT_TAG)
endif

#######################################################################
# Container image release rules. Pre-requirement: Install standard-version
image:
	echo "===> Building container image for $(IMAGE_NAME):$(TAG)"
	docker build --no-cache -t $(PROJECT_NAME)/$(IMAGE_NAME):$(TAG) .

run:
	@echo "===> Run container image for $(PROJECT_NAME)/$(IMAGE_NAME):$(TAG) at port 8081"
	docker run -it -p $(INSTANCE_PORT):$(CONTAINER_PORT) $(PROJECT_NAME)/$(IMAGE_NAME):$(TAG)

bump:
ifeq (,$(findstring $(INIT_TAG), $(REMOTE_TAGS)))
	@standard-version --first-release --releaseCommitMessageFormat "$$RELEASE_COMMIT_MESSAGE"
else
	@standard-version --releaseCommitMessageFormat "$$RELEASE_COMMIT_MESSAGE"
endif