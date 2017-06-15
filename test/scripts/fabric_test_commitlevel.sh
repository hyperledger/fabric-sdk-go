#!/usr/bin/env bash

# This variable can be altered to test against prebuilt fabric and fabric-ca images
# If set to false, CI will build images from scratch for the commit levels specified below
export USE_PREBUILT_IMAGES=true

#file used for automatic integration build test
#This should always match the compatibility specified in the README.md
export FABRIC_COMMIT=v1.0.0-beta
export FABRIC_CA_COMMIT=v1.0.0-beta
