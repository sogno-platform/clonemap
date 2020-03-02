# Developer Guide

## Repository

If you want to implement new features in cloneMAP clone this repository.

## Docker registry

This projects uses the built in Docker registry of GitLab. Stored images are

* registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/ams
* registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/df
* registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/logger
* registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/agency

These images are automatically updated to the latest version by the GitLab CI/CD.
If want to use the same registry please use alternative names like e.g. registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/ams_dev. Don't forget to adjust image names in the Kubernetes yaml file (see below).
