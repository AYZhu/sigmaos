# Building, installing, and running sigmaos.

To make root file system for sigmaos  (needs to be run only once)
# ./install-rootfs.sh

To build sigmaos (and generate a new build version), run:
$ ./make.sh --norace

To upload the built binaries to an s3 bucket corresponding to realm REALM, run
(optional, purely local development also possible as described below):
$ ./upload.sh --realm REALM

To install the sigmaos kernel (realm and kernel packages) for realm REALM, run:
$ ./install.sh --realm REALM --from s3

If running without internet connectivity, everything can be installed locally
by running:
$ ./install.sh --realm REALM --from local

To garbage-collect old build versions from s3 for realm REALM:
$ ./rm-old-versions-s3.sh --realm REALM

To start sigmaos (having already run make.sh, optionally upload.sh, and then
install.sh), and create a realm REALM run:
$ ./start.sh --realm REALM

To stop sigmaos, run:
$ ./stop.sh

To run tests for package PACKAGE_NAME, run:
$ go test -v sigmaos/PACKAGE_NAME

============================================================

Full build flow for 4 development modes, first set the realm name in an
environment variable by running:
$ export REALM_NAME=fkaashoek

Then, run one of the 4 options below:

1. When developing locally and testing locally without internet access (no
access to s3), run:
$ ./make.sh --norace && ./install.sh --realm $REALM_NAME --from local

2. When developing locally and testing locally with internet access (pushing
and pulling from s3) and without garbage collecting old binary versions stored
in s3, run:
$ ./make.sh --norace && ./upload.sh --realm $REALM_NAME && ./install.sh --realm $REALM_NAME --from s3

3. When developing locally and testing locally with internet access (pushing
and pulling from s3), and garbage collecting old binary versions stored in s3,
run:
$ ./make.sh --norace && ./upload.sh --realm $REALM_NAME && ./install.sh --realm $REALM_NAME --from s3 && ./rm-old-versions-s3.sh --realm $REALM_NAME

4. When developing locally and testing on ec2, refer to "aws/README.txt" to
build the binaries on an ec2 instance. Then, run tests as described above.
