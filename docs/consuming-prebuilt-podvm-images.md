# Consuming pre-built podvm images

The project has automation in place to build container images that contain the qcow2 file for the various Linux distributions and cloud providers supported. Those images are re-built for each release of this project.

You can find the images available at https://quay.io/organization/confidential-containers with the *podvm-[CLOUD_PROVIDER]-[DISTRO]* name pattern. For example, https://quay.io/repository/confidential-containers/podvm-libvirt-ubuntu hosts the Ubuntu images for Libvirt.

The easiest way to extract the qcow2 file from the podvm container image is using the [`download-image.sh`](../podvm/hack/download-image.sh) script. For example, to extract the file from the *podvm-libvirt-ubuntu* image:

```
$ cd podvm
$ ./hack/download-image.sh quay.io/confidential-containers/podvm-libvirt-ubuntu . -o podvm.qcow2
```

In case your workload images are pulled from a private registry then you need to provide the authentication file by either [installing along with the cloud-api-adaptor deployment](registries-authentication.md#deploy-authentication-file-along-with-cloud-api-adaptor-deployment) or [statically embedding in the podvm image](registries-authentication.md#statically-embed-authentication-file-in-podvm-image). With the later you will need to build the image from sources, so find detailed instructions in [podvm/README.md](../podvm/README.md).
