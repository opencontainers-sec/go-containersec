# go-containersec
a go security module for container runtime

## Protect CVE-2024-21626
If you want to update runc to 1.1.12, you can choose dmz as the entrypoint of the container:

dmz entrypoint arg0 arg1 ...

## Another way to protect CVE-2019-5736
https://github.com/lifubang/runc/pull/62

To use this similar way to protect CVE-2024-21626, we still have a little work to do, it will be comming soon.
