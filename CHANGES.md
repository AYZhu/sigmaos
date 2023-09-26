# first pass
- added cpython to ./make.sh steps for user programs
- changed sigmaos to depend on ubuntu because alpine does not have the libraries we want

## TODO:
- determine if we should move the building onto alpine
- find a way to cache the python makefile configuration (does not need to be re-done each time)
    - just generally incremental build because dear god this is painful
- how does user-mode python get its libraries?